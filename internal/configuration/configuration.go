package configuration

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"dario.cat/mergo"
	"github.com/logrusorgru/aurora/v4"
	"gopkg.in/yaml.v3"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

// Configuration holds the tool's configuration file contents.
type Configuration struct {
	Version string                  `yaml:"version"`
	Sources []string                `yaml:"sources"`
	Keys    map[string]sources.Keys `yaml:"keys"`
}

// Write writes the configuration to the YAML file at path, creating the
// containing directory with 0750 permissions and the file with 0600.
//
// Parameters:
//   - path (string): the configuration file location.
//
// Returns:
//   - err (error): non-nil on any filesystem or encoding failure.
func (cfg *Configuration) Write(path string) (err error) {
	directory := filepath.Dir(path)
	indentation := 4

	if _, err = os.Stat(directory); errors.Is(err, os.ErrNotExist) {
		if directory != "" {
			if err = os.MkdirAll(directory, 0o750); err != nil {
				return
			}
		}
	}

	var file *os.File

	file, err = os.OpenFile(path, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600) //nolint:gosec // path is the user-specified configuration file location.
	if err != nil {
		return
	}

	defer func() {
		if closeErr := file.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	enc := yaml.NewEncoder(file)

	enc.SetIndent(indentation)

	if err = enc.Encode(&cfg); err != nil {
		return
	}

	if err = enc.Close(); err != nil {
		return
	}

	return
}

// Name holds the tool's name.
const Name = "xsubfind3r"

// Version holds the tool's version. It is overridden at build time via
// -ldflags -X github.com/hueristiq/xsubfind3r/internal/configuration.Version.
var Version = "1.2.0"

// Banner returns the colored application banner, embedding the current
// [Version].
func Banner(au *aurora.Aurora) (banner string) {
	banner = au.Sprintf(
		au.BrightBlue(`
                _      __ _           _ _____
__  _____ _   _| |__  / _(_)_ __   __| |___ / _ __
\ \/ / __| | | | '_ \| |_| | '_ \ / _`+"`"+` | |_ \| '__|
 >  <\__ \ |_| | |_) |  _| | | | | (_| |___) | |
/_/\_\___/\__,_|_.__/|_| |_|_| |_|\__,_|____/|_|
                                             %s`).Bold(),
		au.BrightRed("v"+Version).Bold().Italic(),
	) + "\n\n"

	return
}

// UserConfigDirectoryPath returns the user-specific configuration directory
// (usually `$HOME/.config`).
//
// Returns:
//   - userConfigDirectoryPath (string): the resolved directory path.
//   - err (error): non-nil when the directory cannot be resolved (for
//     example, on a `$HOME`-less system).
func UserConfigDirectoryPath() (userConfigDirectoryPath string, err error) {
	userConfigDirectoryPath, err = os.UserConfigDir()
	if err != nil {
		err = fmt.Errorf("failed to resolve user configuration directory: %w", err)

		return
	}

	return
}

// DefaultConfigurationFilePath returns the default configuration file path.
// When the user configuration directory cannot be resolved, the current
// directory is used as a fallback.
func DefaultConfigurationFilePath() (path string) {
	userConfigDirectoryPath, err := UserConfigDirectoryPath()
	if err != nil {
		userConfigDirectoryPath = "."
	}

	path = filepath.Join(userConfigDirectoryPath, Name, "config.yaml")

	return
}

// DefaultConfiguration holds the configuration written on first run: the
// current version, the full source list, and an empty key map for every
// keyed source.
var DefaultConfiguration = Configuration{
	Version: Version,
	Sources: sources.List,
	Keys: map[string]sources.Keys{
		sources.BEVIGIL:        {},
		sources.BUILTWITH:      {},
		sources.C99:            {},
		sources.CENSYS:         {},
		sources.CERTSPOTTER:    {},
		sources.CHAOS:          {},
		sources.DNSDB:          {},
		sources.DNSDUMPSTER:    {},
		sources.FOFA:           {},
		sources.FULLHUNT:       {},
		sources.GITHUB:         {},
		sources.GITLAB:         {},
		sources.INTELLIGENCEX:  {},
		sources.LEAKIX:         {},
		sources.LEAKRADAR:      {},
		sources.MERKLEMAP:      {},
		sources.NETLAS:         {},
		sources.ONYPHE:         {},
		sources.QUAKE:          {},
		sources.SECURITYTRAILS: {},
		sources.SHODAN:         {},
		sources.URLSCAN:        {},
		sources.VIRUSTOTAL:     {},
		sources.WHOISXMLAPI:    {},
		sources.ZOOMEYE:        {},
	},
}

// CreateOrUpdate creates the configuration file at path when it does not
// exist yet, or refreshes it in place when its version or source list has
// drifted from the defaults.
//
// Parameters:
//   - path (string): the configuration file location.
//
// Returns:
//   - err (error): non-nil on any filesystem or decoding failure.
func CreateOrUpdate(path string) (err error) {
	var cfg Configuration

	_, err = os.Stat(path)

	switch {
	case errors.Is(err, os.ErrNotExist):
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

		if cfg.Version != Version || len(cfg.Sources) != len(sources.List) {
			if err = mergo.Merge(&cfg, DefaultConfiguration); err != nil {
				return
			}

			cfg.Version = Version
			cfg.Sources = sources.List

			if err = cfg.Write(path); err != nil {
				return
			}
		}
	}

	return
}

// Read reads and decodes the configuration file at path.
//
// Parameters:
//   - path (string): the configuration file location.
//
// Returns:
//   - cfg (Configuration): the decoded configuration.
//   - err (error): non-nil on any filesystem or decoding failure.
func Read(path string) (cfg Configuration, err error) {
	file, err := os.Open(path) //nolint:gosec // path is the user-specified configuration file location.
	if err != nil {
		return
	}

	defer func() {
		if closeErr := file.Close(); err == nil && closeErr != nil {
			err = closeErr
		}
	}()

	if err = yaml.NewDecoder(file).Decode(&cfg); err != nil {
		return
	}

	return
}
