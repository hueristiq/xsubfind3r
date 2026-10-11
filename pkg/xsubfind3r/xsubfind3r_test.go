//nolint:testpackage // White-box tests verify per-finder source isolation.
package xsubfind3r

import (
	"sync"
	"testing"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

const testUserAgent = "test"

func TestNewDoesNotShareSourceInstances(t *testing.T) {
	t.Parallel()

	cfg := &Configuration{
		Client:       &ClientConfiguration{UserAgent: testUserAgent},
		SourcesToUse: []string{sources.GITHUB, sources.SECURITYTRAILS},
		Keys: map[string]sources.Keys{
			sources.GITHUB: {"key-1"},
		},
	}

	finder1, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	finder2, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	for _, name := range cfg.SourcesToUse {
		if finder1.sources[name] == finder2.sources[name] {
			t.Fatalf("source %q is shared between finders", name)
		}
	}
}

func TestNewDoesNotMutateConfiguration(t *testing.T) {
	t.Parallel()

	cfg := &Configuration{
		Client: &ClientConfiguration{UserAgent: testUserAgent},
	}

	_, err := New(cfg)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	if cfg.SourcesToUse != nil {
		t.Fatalf("SourcesToUse = %v, want nil (caller configuration must not be mutated)", cfg.SourcesToUse)
	}
}

func TestNewUnknownSource(t *testing.T) {
	t.Parallel()

	cfg := &Configuration{
		Client:       &ClientConfiguration{UserAgent: testUserAgent},
		SourcesToUse: []string{sources.CRTSH, "no-such-source"},
	}

	_, err := New(cfg)
	if err == nil {
		t.Fatal("New() succeeded, want error for unknown source")
	}
}

func TestNewConcurrent(t *testing.T) {
	t.Parallel()

	cfg := &Configuration{
		Client:       &ClientConfiguration{UserAgent: testUserAgent},
		SourcesToUse: []string{sources.GITHUB},
		Keys: map[string]sources.Keys{
			sources.GITHUB: {"key-1"},
		},
	}

	wg := &sync.WaitGroup{}

	for range 20 {
		wg.Add(1)

		go func() {
			defer wg.Done()

			_, err := New(cfg)
			if err != nil {
				t.Errorf("New() error: %v", err)
			}
		}()
	}

	wg.Wait()
}

func TestSourceFactoriesReturnsCopy(t *testing.T) {
	t.Parallel()

	factories := SourceFactories()

	factories["injected"] = nil

	factories2 := SourceFactories()

	if _, ok := factories2["injected"]; ok {
		t.Fatal("SourceFactories() leaked a mutation into the registry")
	}
}
