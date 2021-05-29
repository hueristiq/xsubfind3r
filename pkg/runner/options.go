package runner

import (
	"math/rand"
	"os"
	"path"

	"github.com/signedsecurity/sigsubfind3r/pkg/sources"
	"gopkg.in/yaml.v3"
)

type Configuration struct {
	Version string   `yaml:"version"`
	Sources []string `yaml:"sources"`
	Keys    struct {
		Chaos  []string `yaml:"chaos"`
		GitHub []string `yaml:"github"`
	}
}

type Options struct {
	Domain         string
	SourcesExclude string
	SourcesUse     string

	YAMLConfig Configuration
}

func ParseOptions(options *Options) (*Options, error) {
	directory, err := os.UserHomeDir()
	if err != nil {
		return options, err
	}

	version := "1.0.0"
	configPath := directory + "/.config/sigsubfind3r/conf.yaml"

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		configuration := Configuration{
			Version: version,
			Sources: sources.All,
		}

		directory, _ := path.Split(configPath)

		err := makeDirectory(directory)
		if err != nil {
			return options, err
		}

		err = configuration.MarshalWrite(configPath)
		if err != nil {
			return options, err
		}

		options.YAMLConfig = configuration
	} else {
		configuration, err := UnmarshalRead(configPath)
		if err != nil {
			return options, err
		}

		if configuration.Version != version {
			configuration.Sources = sources.All
			configuration.Version = version

			err := configuration.MarshalWrite(configPath)
			if err != nil {
				return options, err
			}
		}

		options.YAMLConfig = configuration
	}

	return options, nil
}

func makeDirectory(directory string) error {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		if directory != "" {
			err = os.MkdirAll(directory, os.ModePerm)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func (config *Configuration) MarshalWrite(file string) error {
	f, err := os.OpenFile(file, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0755)
	if err != nil {
		return err
	}

	enc := yaml.NewEncoder(f)
	enc.SetIndent(4)
	err = enc.Encode(&config)
	f.Close()
	return err
}

func UnmarshalRead(file string) (Configuration, error) {
	config := Configuration{}

	f, err := os.Open(file)
	if err != nil {
		return config, err
	}

	err = yaml.NewDecoder(f).Decode(&config)

	f.Close()

	return config, err
}

func (config *Configuration) GetKeys() sources.Keys {
	keys := sources.Keys{}

	chaosKeysCount := len(config.Keys.Chaos)
	if chaosKeysCount > 0 {
		keys.Chaos = config.Keys.Chaos[rand.Intn(chaosKeysCount)]
	}

	if len(config.Keys.GitHub) > 0 {
		keys.GitHub = config.Keys.GitHub
	}

	return keys
}
