package main

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"

	"dario.cat/mergo"
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

	domain           string
	listSources      bool
	sourcesToExclude []string
	sourcesToUse     []string
	monochrome       bool
	output           string
	verbosity        string
)

func init() {
	pflag.StringVarP(&domain, "domain", "d", "", "")
	pflag.BoolVarP(&listSources, "sources", "s", false, "")
	pflag.StringSliceVar(&sourcesToExclude, "exclude-sources", []string{}, "")
	pflag.StringSliceVar(&sourcesToUse, "use-sources", []string{}, "")
	pflag.BoolVarP(&monochrome, "monochrome", "m", false, "")
	pflag.StringVarP(&output, "output", "o", "", "")
	pflag.StringVarP(&verbosity, "verbosity", "v", string(levels.LevelInfo), "")

	pflag.CommandLine.SortFlags = false
	pflag.Usage = func() {
		fmt.Fprintln(os.Stderr, configuration.BANNER)

		h := "USAGE:\n"
		h += "  xsubfind3r [OPTIONS]\n"

		h += "\nTARGET:\n"
		h += "  -d, --domain string             target domain\n"

		h += "\nSOURCES:\n"
		h += " -s,  --sources bool              list available sources\n"
		h += "      --exclude-sources strings   comma(,) separated list of sources to exclude\n"
		h += "      --use-sources strings       comma(,) separated list of sources to use\n"

		h += "\nOUTPUT:\n"
		h += "  -m, --monochrome                no colored output mode\n"
		h += "  -o, --output string             output file to write found URLs\n"
		h += fmt.Sprintf("  -v, --verbosity                 debug, info, warning, error, fatal or silent (default: %s)\n\n", string(levels.LevelInfo))

		fmt.Fprintln(os.Stderr, h)
	}

	pflag.Parse()

	// Initialize logger
	hqgolog.DefaultLogger.SetMaxLevel(levels.LevelStr(verbosity))
	hqgolog.DefaultLogger.SetFormatter(formatter.NewCLI(&formatter.CLIOptions{
		Colorize: !monochrome,
	}))

	// Handle configuration on initial run
	var (
		err    error
		config configuration.Configuration
	)

	_, err = os.Stat(configuration.ConfigurationFilePath)
	if err != nil {
		if os.IsNotExist(err) {
			config = configuration.Default

			if err = configuration.Write(&config); err != nil {
				hqgolog.Fatal().Msg(err.Error())
			}
		} else {
			hqgolog.Fatal().Msg(err.Error())
		}
	} else {
		config, err = configuration.Read()
		if err != nil {
			hqgolog.Fatal().Msg(err.Error())
		}

		if config.Version != configuration.VERSION {
			if err = mergo.Merge(&config, configuration.Default); err != nil {
				hqgolog.Fatal().Msg(err.Error())
			}

			config.Version = configuration.VERSION

			if err = configuration.Write(&config); err != nil {
				hqgolog.Fatal().Msg(err.Error())
			}
		}
	}

	au = aurora.NewAurora(!monochrome)
}

func main() {
	if verbosity != string(levels.LevelSilent) {
		fmt.Fprintln(os.Stderr, configuration.BANNER)
	}

	config, err := configuration.Read()
	if err != nil {
		hqgolog.Fatal().Msg(err.Error())
	}

	keys := config.GetKeys()

	// Handle sources listing
	if listSources {
		hqgolog.Info().Msgf("current list of the available %v sources", au.Underline(strconv.Itoa(len(config.Sources))).Bold())
		hqgolog.Info().Msg("sources marked with an * needs key or token")
		hqgolog.Print().Msg("")

		needsKey := make(map[string]interface{})
		keysElem := reflect.ValueOf(&keys).Elem()

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

	// Handle URLs finding
	if verbosity != string(levels.LevelSilent) {
		hqgolog.Info().Msgf("finding subdomains for %v.", au.Underline(domain).Bold())
		hqgolog.Print().Msg("")
	}

	options := &xsubfind3r.Options{
		Domain:           domain,
		SourcesToExclude: sourcesToExclude,
		SourcesToUSe:     sourcesToUse,
		Keys:             keys,
	}

	finder := xsubfind3r.New(options)
	subdomains := finder.Find()

	if output != "" {
		directory := filepath.Dir(output)

		if _, err := os.Stat(directory); os.IsNotExist(err) {
			if err = os.MkdirAll(directory, os.ModePerm); err != nil {
				hqgolog.Fatal().Msg(err.Error())
			}
		}

		file, err := os.OpenFile(output, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			hqgolog.Fatal().Msg(err.Error())
		}

		defer file.Close()

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
		for subdomains := range subdomains {
			if verbosity == string(levels.LevelSilent) {
				hqgolog.Print().Msg(subdomains.Value)
			} else {
				hqgolog.Print().Msgf("[%s] %s", au.BrightBlue(subdomains.Source), subdomains.Value)
			}
		}
	}
}
