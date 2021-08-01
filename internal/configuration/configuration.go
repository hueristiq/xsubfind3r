package configuration

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path"
	"path/filepath"

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

const (
	VERSION string = "1.0.0"
)

var (
	BANNER string = fmt.Sprintf(`
     _                 _      __ _           _ _____
 ___(_) __ _ ___ _   _| |__  / _(_)_ __   __| |___ / _ __
/ __| |/ _`+"`"+` / __| | | | '_ \| |_| | '_ \ / _`+"`"+` | |_ \| '__|
\__ \ | (_| \__ \ |_| | |_) |  _| | | | | (_| |___) | |
|___/_|\__, |___/\__,_|_.__/|_| |_|_| |_|\__,_|____/|_| v%s
       |___/
`, VERSION)

	CONFDIR string = func() (directory string) {
		home, err := os.UserHomeDir()
		if err != nil {
			log.Fatalln(err)
		}

		directory = filepath.Join(home, ".config", "sigsubfind3r")

		return
	}()
)

func (options *Options) Parse() (err error) {
	confYAMLFile := filepath.Join(CONFDIR, "conf.yaml")

	if _, err := os.Stat(confYAMLFile); os.IsNotExist(err) {
		configuration := Configuration{
			Version: VERSION,
			Sources: sources.All,
		}

		directory, _ := path.Split(confYAMLFile)

		err := makeDirectory(directory)
		if err != nil {
			return err
		}

		err = configuration.MarshalWrite(confYAMLFile)
		if err != nil {
			return err
		}

		options.YAMLConfig = configuration
	} else {
		configuration, err := UnmarshalRead(confYAMLFile)
		if err != nil {
			return err
		}

		if configuration.Version != VERSION {
			configuration.Sources = sources.All
			configuration.Version = VERSION

			err := configuration.MarshalWrite(confYAMLFile)
			if err != nil {
				return err
			}
		}

		options.YAMLConfig = configuration
	}

	return
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
