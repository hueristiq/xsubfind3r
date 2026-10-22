package configuration_test

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/hueristiq/xsubfind3r/internal/configuration"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

func TestCreateOrUpdateWritesDefaults(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "nested", "config.yaml")

	if err := configuration.CreateOrUpdate(path); err != nil {
		t.Fatalf("CreateOrUpdate: %v", err)
	}

	cfg, err := configuration.Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if cfg.Version != configuration.Version {
		t.Fatalf("version = %q, want %q", cfg.Version, configuration.Version)
	}

	if len(cfg.Sources) != len(sources.List) {
		t.Fatalf("sources = %d, want %d", len(cfg.Sources), len(sources.List))
	}

	if len(cfg.Keys) != len(configuration.DefaultConfiguration.Keys) {
		t.Fatalf("keys = %d, want %d", len(cfg.Keys), len(configuration.DefaultConfiguration.Keys))
	}

	if runtime.GOOS != "windows" {
		info, statErr := os.Stat(path)
		if statErr != nil {
			t.Fatalf("Stat: %v", statErr)
		}

		if info.Mode().Perm() != 0o600 {
			t.Fatalf("mode = %v, want 0600", info.Mode().Perm())
		}
	}

	entries, err := os.ReadDir(filepath.Dir(path))
	if err != nil {
		t.Fatalf("ReadDir: %v", err)
	}

	if len(entries) != 1 {
		t.Fatalf("directory entries = %d, want 1", len(entries))
	}
}

func TestCreateOrUpdateRecoversEmptyFile(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.yaml")

	if err := os.WriteFile(path, nil, 0o600); err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	if err := configuration.CreateOrUpdate(path); err != nil {
		t.Fatalf("CreateOrUpdate: %v", err)
	}

	cfg, err := configuration.Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if cfg.Version != configuration.Version {
		t.Fatalf("version = %q, want %q", cfg.Version, configuration.Version)
	}

	if len(cfg.Sources) != len(sources.List) {
		t.Fatalf("sources = %d, want %d", len(cfg.Sources), len(sources.List))
	}
}

func TestCreateOrUpdatePreservesUserConfiguration(t *testing.T) {
	t.Parallel()

	path := filepath.Join(t.TempDir(), "config.yaml")

	cfg := configuration.Configuration{
		Version: "old-version",
		Sources: []string{sources.Crtsh},
		Keys: map[string]sources.Keys{
			sources.Shodan: {"custom-key"},
		},
	}

	if err := cfg.Write(path); err != nil {
		t.Fatalf("Write: %v", err)
	}

	if err := configuration.CreateOrUpdate(path); err != nil {
		t.Fatalf("CreateOrUpdate: %v", err)
	}

	got, err := configuration.Read(path)
	if err != nil {
		t.Fatalf("Read: %v", err)
	}

	if len(got.Sources) != 1 || got.Sources[0] != sources.Crtsh {
		t.Fatalf("sources = %v, want [%s]", got.Sources, sources.Crtsh)
	}

	if len(got.Keys[sources.Shodan]) != 1 || got.Keys[sources.Shodan][0] != "custom-key" {
		t.Fatalf("shodan key = %v, want [custom-key]", got.Keys[sources.Shodan])
	}

	if got.Version != configuration.Version {
		t.Fatalf("version = %q, want %q", got.Version, configuration.Version)
	}

	if len(got.Keys) != len(configuration.DefaultConfiguration.Keys) {
		t.Fatalf("keys = %d, want %d", len(got.Keys), len(configuration.DefaultConfiguration.Keys))
	}
}
