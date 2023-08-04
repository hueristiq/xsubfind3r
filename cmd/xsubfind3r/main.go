package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"github.com/hueristiq/hqgolog"
	"github.com/hueristiq/hqgolog/formatter"
	"github.com/hueristiq/hqgolog/levels"
	"github.com/hueristiq/xsubfind3r/internal/configuration"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/logrusorgru/aurora/v3"
	"github.com/spf13/pflag"
)

var (
	au aurora.Aurora

	domains             []string
	domainsListFilePath string
	listSources         bool
	sourcesToUse        []string
	sourcesToExclude    []string
	monochrome          bool
	output              string
	outputDirectory     string
	verbosity           string
	YAMLConfigFile      string
)

func init() {
	// defaults
	defaultYAMLConfigFile := fmt.Sprintf("~/.hueristiq/%s/config.yaml", configuration.NAME)

	// Handle CLI arguments, flags & help message (pflag)
	pflag.StringSliceVarP(&domains, "domain", "d", []string{}, "")
	pflag.StringVarP(&domainsListFilePath, "list", "l", "", "")
	pflag.BoolVar(&listSources, "sources", false, "")
	pflag.StringSliceVarP(&sourcesToUse, "use-sources", "u", []string{}, "")
	pflag.StringSliceVarP(&sourcesToExclude, "exclude-sources", "e", []string{}, "")
	pflag.BoolVar(&monochrome, "no-color", false, "")
	pflag.StringVarP(&output, "output", "o", "", "")
	pflag.StringVarP(&outputDirectory, "outputDirectory", "O", "", "")
	pflag.StringVarP(&verbosity, "verbosity", "v", string(levels.LevelInfo), "")
	pflag.StringVarP(&YAMLConfigFile, "configuration", "c", defaultYAMLConfigFile, "")

	pflag.CommandLine.SortFlags = false
	pflag.Usage = func() {
		fmt.Fprintln(os.Stderr, configuration.BANNER)

		h := "USAGE:\n"
		h += "  xsubfind3r [OPTIONS]\n"

		h += "\nINPUT:\n"
		h += " -d, --domain string[]                 target domains\n"
		h += " -l, --list string                     target domains list file path\n"

		h += "\nSOURCES:\n"
		h += "      --sources bool                   list supported sources\n"
		h += " -u,  --sources-to-use string[]        comma(,) separeted sources to use\n"
		h += " -e,  --sources-to-exclude string[]    comma(,) separeted sources to exclude\n"

		h += "\nOUTPUT:\n"
		h += "     --no-color bool                   disable colored output\n"
		h += " -o, --output string                   output subdomains file path\n"
		h += " -O, --output-directory string         output subdomains directory path\n"
		h += fmt.Sprintf(" -v, --verbosity string                debug, info, warning, error, fatal or silent (default: %s)\n", string(levels.LevelInfo))

		h += "\nCONFIGURATION:\n"
		h += fmt.Sprintf(" -c,  --configuration string           configuration file path (default: %s)\n", defaultYAMLConfigFile)

		fmt.Fprintln(os.Stderr, h)
	}

	pflag.Parse()

	// Initialize logger (hqgolog)
	hqgolog.DefaultLogger.SetMaxLevel(levels.LevelStr(verbosity))
	hqgolog.DefaultLogger.SetFormatter(formatter.NewCLI(&formatter.CLIOptions{
		Colorize: !monochrome,
	}))

	// Create | Update configuration
	if strings.HasPrefix(YAMLConfigFile, "~") {
		home, err := os.UserHomeDir()
		if err != nil {
			hqgolog.Fatal().Msg(err.Error())
		}

		YAMLConfigFile = strings.Replace(YAMLConfigFile, "~", home, 1)
	}

	if err := configuration.CreateUpdate(YAMLConfigFile); err != nil {
		hqgolog.Fatal().Msg(err.Error())
	}

	au = aurora.NewAurora(!monochrome)
}

func main() {
	// Print banner.
	if verbosity != string(levels.LevelSilent) {
		fmt.Fprintln(os.Stderr, configuration.BANNER)
	}

	var err error

	var config configuration.Configuration

	// Read in configuration.
	config, err = configuration.Read(YAMLConfigFile)
	if err != nil {
		hqgolog.Fatal().Msg(err.Error())
	}

	// If --sources: List suported sources & exit.
	if listSources {
		hqgolog.Info().Msgf("listing, %v, current supported sources.", au.Underline(strconv.Itoa(len(config.Sources))).Bold())
		hqgolog.Info().Msgf("sources marked with %v take in key(s) or token(s).", au.Underline("*").Bold())
		hqgolog.Print().Msg("")

		needsKey := make(map[string]interface{})
		keysElem := reflect.ValueOf(&config.Keys).Elem()

		for i := 0; i < keysElem.NumField(); i++ {
			needsKey[strings.ToLower(keysElem.Type().Field(i).Name)] = keysElem.Field(i).Interface()
		}

		for _, source := range config.Sources {
			if _, ok := needsKey[source]; ok {
				hqgolog.Print().Msgf("> %s *", source)
			} else {
				hqgolog.Print().Msgf("> %s", source)
			}
		}

		hqgolog.Print().Msg("")

		os.Exit(0)
	}

	// Load input domains.

	// input domains: file
	if domainsListFilePath != "" {
		var file *os.File

		file, err = os.Open(domainsListFilePath)
		if err != nil {
			hqgolog.Error().Msg(err.Error())
		}

		scanner := bufio.NewScanner(file)

		for scanner.Scan() {
			domain := scanner.Text()

			if domain != "" {
				domains = append(domains, domain)
			}
		}

		if err = scanner.Err(); err != nil {
			hqgolog.Error().Msg(err.Error())
		}
	}

	// input domains: stdin
	if hasStdin() {
		scanner := bufio.NewScanner(os.Stdin)

		for scanner.Scan() {
			domain := scanner.Text()

			if domain != "" {
				domains = append(domains, domain)
			}
		}

		if err = scanner.Err(); err != nil {
			hqgolog.Error().Msg(err.Error())
		}
	}

	// Find and output subdomains.
	options := &xsubfind3r.Options{
		SourcesToExclude: sourcesToExclude,
		SourcesToUSe:     sourcesToUse,
		Keys:             config.Keys,
	}

	finder := xsubfind3r.New(options)

	var consolidatedWriter *bufio.Writer

	if output != "" {
		directory := filepath.Dir(output)

		mkdir(directory)

		var consolidatedFile *os.File

		consolidatedFile, err = os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			hqgolog.Fatal().Msg(err.Error())
		}

		defer consolidatedFile.Close()

		consolidatedWriter = bufio.NewWriter(consolidatedFile)
	}

	if outputDirectory != "" {
		mkdir(outputDirectory)
	}

	for _, domain := range domains {
		subdomains := finder.Find(domain)

		switch {
		case output != "":
			processSubdomains(consolidatedWriter, subdomains, verbosity)
		case outputDirectory != "":
			var domainFile *os.File

			domainFile, err = os.OpenFile(filepath.Join(outputDirectory, domain+".txt"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				hqgolog.Fatal().Msg(err.Error())
			}

			domainWriter := bufio.NewWriter(domainFile)

			processSubdomains(domainWriter, subdomains, verbosity)
		default:
			processSubdomains(nil, subdomains, verbosity)
		}
	}
}

func hasStdin() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}

	isPipedFromChrDev := (stat.Mode() & os.ModeCharDevice) == 0
	isPipedFromFIFO := (stat.Mode() & os.ModeNamedPipe) != 0

	return isPipedFromChrDev || isPipedFromFIFO
}

func mkdir(path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		if err = os.MkdirAll(path, os.ModePerm); err != nil {
			hqgolog.Fatal().Msg(err.Error())
		}
	}
}

func processSubdomains(writer *bufio.Writer, subdomains chan sources.Result, verbosity string) {
	for subdomain := range subdomains {
		switch subdomain.Type {
		case sources.Error:
			hqgolog.Warn().Msgf("Could not run source %s: %s\n", subdomain.Source, subdomain.Error)
		case sources.Subdomain:
			if verbosity == string(levels.LevelDebug) {
				hqgolog.Print().Msgf("[%s] %s", au.BrightBlue(subdomain.Source), subdomain.Value)
			} else {
				hqgolog.Print().Msg(subdomain.Value)
			}

			if writer != nil {
				fmt.Fprintln(writer, subdomain.Value)

				if err := writer.Flush(); err != nil {
					hqgolog.Fatal().Msg(err.Error())
				}
			}
		}
	}
}
