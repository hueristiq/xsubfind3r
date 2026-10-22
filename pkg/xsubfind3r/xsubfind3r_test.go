//nolint:testpackage // white-box tests cover unexported validation and relay helpers.
package xsubfind3r

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

const (
	exampleDomain = "example.com"
	subdomainA    = "a.example.com"
	subdomainB    = "b.example.com"
)

func TestValidateSubdomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		target string
		want   string
		valid  bool
	}{
		{name: "plain", target: subdomainA, want: subdomainA, valid: true},
		{name: "uppercase", target: "A.Example.COM", want: subdomainA, valid: true},
		{name: "trailing dot", target: subdomainA + ".", want: subdomainA, valid: true},
		{name: "surrounding whitespace", target: " " + subdomainA + " ", want: subdomainA, valid: true},
		{name: "wildcard", target: "*." + subdomainA, want: subdomainA, valid: true},
		{name: "two f label", target: "2f.example.com", want: "2f.example.com", valid: true},
		{name: "two f a label", target: "2fa.example.com", want: "2fa.example.com", valid: true},
		{name: "underscore", target: "a_b.example.com", want: "a_b.example.com", valid: true},
		{name: "encoded slash", target: "%2f.example.com", want: "%2f.example.com", valid: false},
		{name: "wildcard only", target: "*.example.com", want: "example.com", valid: false},
		{name: "leading hyphen", target: "-a.example.com", want: "-a.example.com", valid: false},
		{name: "trailing hyphen", target: "a-.example.com", want: "a-.example.com", valid: false},
		{name: "empty label", target: "a..example.com", want: "a..example.com", valid: false},
		{name: "control byte", target: "a\x00.example.com", want: "a\x00.example.com", valid: false},
		{name: "evil suffix", target: "evil-example.com", want: "evil-example.com", valid: false},
		{name: "suffix trap", target: "example.com.evil.com", want: "example.com.evil.com", valid: false},
		{name: "overlong label", target: strings.Repeat("a", 64) + ".example.com", want: strings.Repeat("a", 64) + ".example.com", valid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got, valid := validateSubdomain(test.target, exampleDomain)

			if got != test.want || valid != test.valid {
				t.Fatalf("validateSubdomain(%q) = (%q, %v), want (%q, %v)", test.target, got, valid, test.want, test.valid)
			}
		})
	}
}

func TestValidHostname(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		hostname string
		valid    bool
	}{
		{name: "empty", hostname: "", valid: false},
		{name: "single label", hostname: "localhost", valid: true},
		{name: "63 byte label", hostname: strings.Repeat("a", 63) + ".com", valid: true},
		{name: "64 byte label", hostname: strings.Repeat("a", 64) + ".com", valid: false},
		{name: "253 byte name", hostname: strings.Join([]string{strings.Repeat("a", 63), strings.Repeat("b", 63), strings.Repeat("c", 63), strings.Repeat("d", 61)}, "."), valid: true},
		{name: "254 byte name", hostname: strings.Join([]string{strings.Repeat("a", 63), strings.Repeat("b", 63), strings.Repeat("c", 63), strings.Repeat("d", 62)}, "."), valid: false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if valid := validHostname(test.hostname); valid != test.valid {
				t.Fatalf("validHostname(len=%d) = %v, want %v", len(test.hostname), valid, test.valid)
			}
		})
	}
}

func TestSourceRegistry(t *testing.T) {
	t.Parallel()

	for _, name := range sources.List {
		t.Run(name, func(t *testing.T) {
			t.Parallel()

			source, ok := NewSource(name)
			if !ok {
				t.Fatalf("NewSource(%q) not registered", name)
			}

			if got := source.Name(); got != name {
				t.Fatalf("source.Name() = %q, want %q", got, name)
			}

			source.UseKeys("test-key")
		})
	}
}

func TestNew(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		cfg     *Configuration
		wantErr bool
	}{
		{name: "nil", cfg: nil, wantErr: true},
		{name: "unknown source", cfg: &Configuration{SourcesToUse: []string{"nope"}}, wantErr: true},
		{name: "unknown exclude", cfg: &Configuration{SourcesToUse: []string{sources.Crtsh}, SourcesToExclude: []string{"nope"}}, wantErr: true},
		{name: "exclude all", cfg: &Configuration{SourcesToUse: []string{sources.Crtsh}, SourcesToExclude: []string{sources.Crtsh}}, wantErr: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			_, err := New(test.cfg)

			if test.wantErr && err == nil {
				t.Fatal("New returned no error")
			}

			if !test.wantErr && err != nil {
				t.Fatalf("New: %v", err)
			}
		})
	}
}

func TestNewSelectsSources(t *testing.T) {
	t.Parallel()

	finder, err := New(&Configuration{
		SourcesToUse:     []string{sources.Crtsh, sources.Wayback},
		SourcesToExclude: []string{sources.Crtsh},
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}

	if len(finder.sources) != 1 {
		t.Fatalf("sources = %d, want 1", len(finder.sources))
	}

	if _, exists := finder.sources[sources.Wayback]; !exists {
		t.Fatalf("sources = %v, want %s", finder.sources, sources.Wayback)
	}
}

func TestFindDeduplicatesAndForwardsErrors(t *testing.T) {
	t.Parallel()

	boom := errors.New("boom")

	finder := &Finder{
		sources: map[string]sources.Source{
			"one": &staticSource{
				name: "one",
				results: []sources.Result{
					sources.SubdomainResult("one", subdomainA),
					sources.SubdomainResult("one", subdomainA),
					sources.WarningResult("one", "capped"),
					sources.ErrorResult("one", boom),
				},
			},
			"two": &staticSource{
				name: "two",
				results: []sources.Result{
					sources.SubdomainResult("two", subdomainB),
					sources.SubdomainResult("two", subdomainA),
				},
			},
		},
	}

	values, warnings, errs := collectResults(t, finder.Find(t.Context(), exampleDomain))

	if len(values) != 2 {
		t.Fatalf("values = %v, want 2", values)
	}

	if warnings != 1 {
		t.Fatalf("warnings = %d, want 1", warnings)
	}

	if !contains(values, subdomainA) || !contains(values, subdomainB) {
		t.Fatalf("values = %v", values)
	}

	if len(errs) != 1 || !errors.Is(errs[0], boom) {
		t.Fatalf("errors = %v, want [boom]", errs)
	}
}

func TestFindCancellationDrainsSources(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()

	finder := &Finder{
		sources: map[string]sources.Source{
			"blocking": &blockingSource{},
		},
	}

	results := finder.Find(ctx, exampleDomain)

	select {
	case <-results:
	case <-time.After(time.Second):
		t.Fatal("no result before cancellation")
	}

	cancel()

	done := make(chan struct{})

	go func() {
		for range results {
		}

		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("results channel did not close after cancellation")
	}
}

func collectResults(t *testing.T, results <-chan sources.Result) (values []string, warnings int, errs []error) {
	t.Helper()

	for result := range results {
		switch result.Type {
		case sources.ResultSubdomain:
			values = append(values, result.Value)
		case sources.ResultError:
			errs = append(errs, result.Error)
		case sources.ResultWarning:
			warnings++
		}
	}

	return
}

func contains(values []string, want string) (found bool) {
	for _, value := range values {
		if value == want {
			return true
		}
	}

	return
}

type staticSource struct {
	name    string
	results []sources.Result
}

func (source *staticSource) Name() (name string) {
	return source.name
}

func (source *staticSource) NeedsKeys() (needs bool) {
	return false
}

func (source *staticSource) UseKeys(keys ...string) {}

func (source *staticSource) Run(ctx context.Context, cfg *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		for _, result := range source.results {
			select {
			case results <- result:
			case <-ctx.Done():
				return
			}
		}
	}()

	return results
}

type blockingSource struct{}

func (source *blockingSource) Name() (name string) {
	return "blocking"
}

func (source *blockingSource) NeedsKeys() (needs bool) {
	return false
}

func (source *blockingSource) UseKeys(keys ...string) {}

func (source *blockingSource) Run(ctx context.Context, cfg *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		for index := range 100 {
			results <- sources.SubdomainResult("blocking", fmt.Sprintf("s%d.%s", index, exampleDomain))
		}
	}()

	return results
}

var (
	_ sources.Source = (*staticSource)(nil)
	_ sources.Source = (*blockingSource)(nil)
)
