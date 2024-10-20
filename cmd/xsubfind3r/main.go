package main

import (
	"bufio"
	"fmt"
	"log"
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
	"github.com/spf13/viper"
)

var (
	au aurora.Aurora

	configurationFilePath string
	domains               []string
	domainsListFilePath   string
	listSources           bool
	sourcesToUse          []string
	sourcesToExclude      []string
	monochrome            bool
	output                string
	outputDirectory       string
	silent                bool
	verbose               bool
)

func init() {
	// Handle CLI arguments, flags & help message (pflag)
	pflag.StringVarP(&configurationFilePath, "configuration", "c", configuration.ConfigurationFilePath, "")
	pflag.StringSliceVarP(&domains, "domain", "d", []string{}, "")
	pflag.StringVarP(&domainsListFilePath, "list", "l", "", "")
	pflag.BoolVar(&listSources, "sources", false, "")
	pflag.StringSliceVarP(&sourcesToUse, "use-sources", "u", []string{}, "")
	pflag.StringSliceVarP(&sourcesToExclude, "exclude-sources", "e", []string{}, "")
	pflag.BoolVar(&monochrome, "monochrome", false, "")
	pflag.StringVarP(&output, "output", "o", "", "")
	pflag.StringVarP(&outputDirectory, "output-directory", "O", "", "")
	pflag.BoolVarP(&silent, "silent", "s", false, "")
	pflag.BoolVarP(&verbose, "verbose", "v", false, "")

	pflag.CommandLine.SortFlags = false
	pflag.Usage = func() {
		fmt.Fprintln(os.Stderr, configuration.BANNER)

		h := "\nUSAGE:\n"
		h += fmt.Sprintf(" %s [OPTIONS]\n", configuration.NAME)

		h += "\nCONFIGURATION:\n"
		defaultConfigurationFilePath := strings.ReplaceAll(configuration.ConfigurationFilePath, configuration.UserDotConfigDirectoryPath, "$HOME/.config")
		h += fmt.Sprintf(" -c, --configuration string            configuration file (default: %s)\n", defaultConfigurationFilePath)

		h += "\nINPUT:\n"
		h += " -d, --domain string[]                 target domain\n"
		h += " -l, --list string                     target domains list file path\n"

		h += "\nTIP: For multiple input domains use comma(,) separated value with `-d`,\n"
		h += "     specify multiple `-d`, load from file with `-l` or load from stdin.\n"

		h += "\nSOURCES:\n"
		h += "     --sources bool                    list supported sources\n"
		h += " -u, --sources-to-use string[]         comma(,) separated sources to use\n"
		h += " -e, --sources-to-exclude string[]     comma(,) separated sources to exclude\n"

		h += "\nOUTPUT:\n"
		h += "     --monochrome bool                 display no color output\n"
		h += " -o, --output string                   output subdomains file path\n"
		h += " -O, --output-directory string         output subdomains directory path\n"
		h += " -s, --silent bool                     display output subdomains only\n"
		h += " -v, --verbose bool                    display verbose output\n"

		fmt.Fprintln(os.Stderr, h)
	}

	pflag.Parse()

	// Initialize configuration management (...with viper)
	if err := configuration.CreateUpdate(configurationFilePath); err != nil {
		hqgolog.Fatal().Msg(err.Error())
	}

	viper.SetConfigFile(configurationFilePath)
	viper.AutomaticEnv()
	viper.SetEnvPrefix("XSUBFIND3R")
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalln(err)
	}

	// Initialize logger (hqgolog)
	hqgolog.DefaultLogger.SetMaxLevel(levels.LevelInfo)

	if verbose {
		hqgolog.DefaultLogger.SetMaxLevel(levels.LevelDebug)
	}

	hqgolog.DefaultLogger.SetFormatter(formatter.NewCLI(&formatter.CLIOptions{
		Colorize: !monochrome,
	}))

	au = aurora.NewAurora(!monochrome)
}

func main() {
	// print banner.
	if !silent {
		fmt.Fprintln(os.Stderr, configuration.BANNER)
	}

	var err error

	var config *configuration.Configuration

	if err := viper.Unmarshal(&config); err != nil {
		hqgolog.Fatal().Msg(err.Error())
	}

	// if `--sources`: List suported sources & exit.
	if listSources {
		hqgolog.Print().Msg("")
		hqgolog.Info().Msgf("listing, %v, current supported sources.", au.Underline(strconv.Itoa(len(config.Sources))).Bold())
		hqgolog.Info().Msgf("sources marked with %v take in key(s) or token(s).", au.Underline("*").Bold())
		hqgolog.Print().Msg("")

		needsKey := make(map[string]interface{})
		keysElem := reflect.ValueOf(&config.Keys).Elem()

		for i := range keysElem.NumField() {
			needsKey[strings.ToLower(keysElem.Type().Field(i).Name)] = keysElem.Field(i).Interface()
		}

		for index := range config.Sources {
			source := config.Sources[index]

			if _, ok := needsKey[source]; ok {
				hqgolog.Print().Msgf("> %s *", source)
			} else {
				hqgolog.Print().Msgf("> %s", source)
			}
		}

		hqgolog.Print().Msg("")

		os.Exit(0)
	}

	// load input domains from file
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

	// load input domains from stdin
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

	// scrape and output subdomains.
	cfg := &xsubfind3r.Configuration{
		SourcesToUSe:     sourcesToUse,
		SourcesToExclude: sourcesToExclude,
		Keys:             config.Keys,
	}

	var finder *xsubfind3r.Finder

	finder, err = xsubfind3r.New(cfg)
	if err != nil {
		hqgolog.Error().Msg(err.Error())

		return
	}

	var consolidatedWriter *bufio.Writer

	if output != "" {
		directory := filepath.Dir(output)

		mkdir(directory)

		var consolidatedFile *os.File

		consolidatedFile, err = os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			hqgolog.Fatal().Msg(err.Error())
		}

		defer consolidatedFile.Close()

		consolidatedWriter = bufio.NewWriter(consolidatedFile)
	}

	if outputDirectory != "" {
		mkdir(outputDirectory)
	}

	for index := range domains {
		domain := domains[index]

		if !silent {
			hqgolog.Print().Msg("")
			hqgolog.Info().Msgf("Finding subdomains for %v...", au.Underline(domain).Bold())
			hqgolog.Print().Msg("")
		}

		subdomains := finder.Find(domain)

		switch {
		case output != "":
			processSubdomains(consolidatedWriter, subdomains)
		case outputDirectory != "":
			var domainFile *os.File

			domainFile, err = os.OpenFile(filepath.Join(outputDirectory, domain+".txt"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
			if err != nil {
				hqgolog.Fatal().Msg(err.Error())
			}

			domainWriter := bufio.NewWriter(domainFile)

			processSubdomains(domainWriter, subdomains)
		default:
			processSubdomains(nil, subdomains)
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

func processSubdomains(writer *bufio.Writer, subdomains chan sources.Result) {
	for subdomain := range subdomains {
		switch subdomain.Type {
		case sources.ResultError:
			if verbose {
				hqgolog.Error().Msgf("%s: %s\n", subdomain.Source, subdomain.Error)
			}
		case sources.ResultSubdomain:
			if verbose {
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
