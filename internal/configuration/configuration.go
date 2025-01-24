package configuration

import (
	"fmt"
	"os"
	"path/filepath"

	"dario.cat/mergo"
	"github.com/hueristiq/hqgolog"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"gopkg.in/yaml.v3"
)

type Configuration struct {
	Version string       `yaml:"version"`
	Sources []string     `yaml:"sources"`
	Keys    sources.Keys `yaml:"keys"`
}

func (cfg *Configuration) Write(path string) (err error) {
	var file *os.File

	directory := filepath.Dir(path)
	identation := 4

	if _, err = os.Stat(directory); os.IsNotExist(err) {
		if directory != "" {
			if err = os.MkdirAll(directory, os.ModePerm); err != nil {
				return
			}
		}
	}

	file, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o755)
	if err != nil {
		return
	}

	defer file.Close()

	enc := yaml.NewEncoder(file)
	enc.SetIndent(identation)
	err = enc.Encode(&cfg)

	return
}

const (
	NAME    = "xsubfind3r"
	VERSION = "0.9.1"
)

var (
	BANNER = fmt.Sprintf(`
       ____        _     _____ _           _ _____
__  __/ ___| _   _| |__ |  ___(_)_ __   __| |___ / _ __
\ \/ /\___ \| | | | '_ \| |_  | | '_ \ / _`+"`"+` | |_ \| '__|
 >  <  ___) | |_| | |_) |  _| | | | | | (_| |___) | |
/_/\_\|____/ \__,_|_.__/|_|   |_|_| |_|\__,_|____/|_|
                                                  v%s

                               Hueristiq (hueristiq.com)
`, VERSION)
	UserDotConfigDirectoryPath = func() (userDotConfig string) {
		var err error

		userDotConfig, err = os.UserConfigDir()
		if err != nil {
			hqgolog.Fatal().Msg(err.Error())
		}

		return
	}()
	DefaultConfigurationFilePath = filepath.Join(UserDotConfigDirectoryPath, NAME, "config.yaml")
	DefaultConfiguration         = Configuration{
		Version: VERSION,
		Sources: sources.List,
		Keys: sources.Keys{
			Bevigil:        []string{},
			BuiltWith:      []string{},
			Censys:         []string{},
			Chaos:          []string{},
			Fullhunt:       []string{},
			GitHub:         []string{},
			Intelx:         []string{},
			SecurityTrails: []string{},
			Shodan:         []string{},
			URLScan:        []string{},
			VirusTotal:     []string{},
		},
	}
)

func CreateOrUpdate(path string) (err error) {
	var cfg Configuration

	_, err = os.Stat(path)

	switch {
	case err != nil && os.IsNotExist(err):
		cfg = DefaultConfiguration

		if err = cfg.Write(path); err != nil {
			return
		}
	case err != nil:
		return
	default:
		cfg, err = Read(path)
		if err != nil {
			return
		}

		if cfg.Version != VERSION || len(cfg.Sources) != len(sources.List) {
			if err = mergo.Merge(&cfg, DefaultConfiguration); err != nil {
				return
			}

			cfg.Version = VERSION
			cfg.Sources = sources.List

			if err = cfg.Write(path); err != nil {
				return
			}
		}
	}

	return
}

func Read(path string) (cfg Configuration, err error) {
	var file *os.File

	file, err = os.Open(path)
	if err != nil {
		return
	}

	defer file.Close()

	if err = yaml.NewDecoder(file).Decode(&cfg); err != nil {
		return
	}

	return
}
