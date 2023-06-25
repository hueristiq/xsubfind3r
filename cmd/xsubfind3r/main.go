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
	"github.com/logrusorgru/aurora/v3"
	"github.com/spf13/pflag"
)

var (
	au aurora.Aurora

	YAMLConfigFile string

	domain string

	sourcesToExclude []string
	listSources      bool
	sourcesToUse     []string

	monochrome bool
	output     string
	verbosity  string
)

func init() {
	// defaults
	defaultYAMLConfigFile := "~/.hueristiq/xsubfind3r/config.yaml"

	// Handle CLI arguments, flags & help message (pflag)
	pflag.StringVarP(&domain, "domain", "d", "", "")

	pflag.StringSliceVarP(&sourcesToExclude, "exclude-sources", "e", []string{}, "")
	pflag.BoolVarP(&listSources, "sources", "s", false, "")
	pflag.StringSliceVarP(&sourcesToUse, "use-sources", "u", []string{}, "")

	pflag.BoolVar(&monochrome, "no-color", false, "")
	pflag.StringVarP(&output, "output", "o", "", "")
	pflag.StringVarP(&verbosity, "verbosity", "v", string(levels.LevelInfo), "")

	pflag.StringVarP(&YAMLConfigFile, "configuration", "c", defaultYAMLConfigFile, "")

	pflag.CommandLine.SortFlags = false
	pflag.Usage = func() {
		fmt.Fprintln(os.Stderr, configuration.BANNER)

		h := "USAGE:\n"
		h += "  xsubfind3r [OPTIONS]\n"

		h += "\nTARGET:\n"
		h += " -d, --domain string              target domain\n"

		h += "\nSOURCES:\n"
		h += " -e,  --exclude-sources string    sources to exclude\n"
		h += " -s,  --sources bool              list sources\n"
		h += " -u,  --use-sources string        sources to use\n"

		h += "\nOUTPUT:\n"
		h += "     --no-color bool              no colored mode\n"
		h += " -o, --output string              output subdomains file path\n"
		h += fmt.Sprintf(" -v, --verbosity string           debug, info, warning, error, fatal or silent (default: %s)\n", string(levels.LevelInfo))

		h += "\nCONFIGURATION:\n"
		h += fmt.Sprintf(" -c,  --configuration string      configuration file path (default: %s)\n", defaultYAMLConfigFile)

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
	// Print Banner
	if verbosity != string(levels.LevelSilent) {
		fmt.Fprintln(os.Stderr, configuration.BANNER)
	}

	// Read in configuration
	config, err := configuration.Read(YAMLConfigFile)
	if err != nil {
		hqgolog.Fatal().Msg(err.Error())
	}

	// List suported sources
	if listSources {
		hqgolog.Info().Msgf("listing %v current supported sources", au.Underline(strconv.Itoa(len(config.Sources))).Bold())
		hqgolog.Info().Msgf("sources with %v needs a key or token", au.Underline("*").Bold())
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

	// Find subdomains
	if verbosity != string(levels.LevelSilent) {
		hqgolog.Info().Msgf("finding subdomains for %v.", au.Underline(domain).Bold())
		hqgolog.Print().Msg("")
	}

	options := &xsubfind3r.Options{
		Domain:           domain,
		SourcesToExclude: sourcesToExclude,
		SourcesToUSe:     sourcesToUse,
		Keys:             config.Keys,
	}

	finder := xsubfind3r.New(options)
	subdomains := finder.Find()

	if output != "" {
		// Create output file path directory
		directory := filepath.Dir(output)

		if _, err := os.Stat(directory); os.IsNotExist(err) {
			if err = os.MkdirAll(directory, os.ModePerm); err != nil {
				hqgolog.Fatal().Msg(err.Error())
			}
		}

		// Create output file
		file, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			hqgolog.Fatal().Msg(err.Error())
		}

		defer file.Close()

		// Write subdomains output file and print on screen
		writer := bufio.NewWriter(file)

		for subdomains := range subdomains {
			if verbosity == string(levels.LevelSilent) {
				hqgolog.Print().Msg(subdomains.Value)
			} else {
				hqgolog.Print().Msgf("[%s] %s", au.BrightBlue(subdomains.Source), subdomains.Value)
			}

			fmt.Fprintln(writer, subdomains.Value)
		}

		if err = writer.Flush(); err != nil {
			hqgolog.Fatal().Msg(err.Error())
		}
	} else {
		// Print subdomains on screen
		for subdomains := range subdomains {
			if verbosity == string(levels.LevelSilent) {
				hqgolog.Print().Msg(subdomains.Value)
			} else {
				hqgolog.Print().Msgf("[%s] %s", au.BrightBlue(subdomains.Source), subdomains.Value)
			}
		}
	}
}
