package configuration

import (
	"os"
	"path/filepath"

	"github.com/hueristiq/hqgolog"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/logrusorgru/aurora/v3"
	"gopkg.in/yaml.v3"
)

type Configuration struct {
	Version string       `yaml:"version"`
	Sources []string     `yaml:"sources"`
	Keys    sources.Keys `yaml:"keys"`
}

type Options struct {
	Domain         string
	SourcesExclude string
	SourcesUse     string

	YAMLConfig Configuration
}

const (
	NAME        string = "xsubfind3r"
	VERSION     string = "0.0.0"
	DESCRIPTION string = "A CLI utility to find domain's known subdomains."
)

var (
	BANNER = aurora.Sprintf(
		aurora.BrightBlue(`
                _      __ _           _ _____      
__  _____ _   _| |__  / _(_)_ __   __| |___ / _ __ 
\ \/ / __| | | | '_ \| |_| | '_ \ / _`+"`"+` | |_ \| '__|
 >  <\__ \ |_| | |_) |  _| | | | | (_| |___) | |   
/_/\_\___/\__,_|_.__/|_| |_|_| |_|\__,_|____/|_| %s

%s
`).Bold(),
		aurora.BrightYellow("v"+VERSION).Bold(),
		aurora.BrightGreen(DESCRIPTION).Italic(),
	)
	rootDirectoryName        = ".hueristiq"
	projectRootDirectoryName = NAME
	ProjectRootDirectoryPath = func(rootDirectoryName, projectRootDirectoryName string) string {
		home, err := os.UserHomeDir()
		if err != nil {
			hqgolog.Fatal().Msg(err.Error())
		}

		return filepath.Join(home, rootDirectoryName, projectRootDirectoryName)
	}(rootDirectoryName, projectRootDirectoryName)
	configurationFileName = "config.yaml"
	ConfigurationFilePath = filepath.Join(ProjectRootDirectoryPath, configurationFileName)
	Default               = Configuration{
		Version: VERSION,
		Sources: sources.List,
		Keys: sources.Keys{
			Chaos:   []string{},
			GitHub:  []string{},
			Intelx:  []string{},
			URLScan: []string{},
		},
	}
)

func Read() (configuration Configuration, err error) {
	var (
		file *os.File
	)

	file, err = os.Open(ConfigurationFilePath)
	if err != nil {
		return
	}

	defer file.Close()

	err = yaml.NewDecoder(file).Decode(&configuration)

	return
}

func Write(configuration *Configuration) (err error) {
	var (
		file       *os.File
		identation = 4
	)

	directory := filepath.Dir(ConfigurationFilePath)

	if _, err = os.Stat(directory); os.IsNotExist(err) {
		if directory != "" {
			if err = os.MkdirAll(directory, os.ModePerm); err != nil {
				return
			}
		}
	}

	file, err = os.OpenFile(ConfigurationFilePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return
	}

	defer file.Close()

	enc := yaml.NewEncoder(file)
	enc.SetIndent(identation)
	err = enc.Encode(&configuration)

	return
}
