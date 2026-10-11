package securitytrails_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/securitytrails"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/sourcetest"
)

func TestSourceRun(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Apikey"); got != "test-key" {
			t.Errorf("APIKEY header = %q, want %q", got, "test-key")
		}

		if r.URL.Path != "/v1/domain/example.com/subdomains" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}

		_, _ = w.Write([]byte(`{
			"subdomain_count": 3,
			"subdomains": ["api", "www.", "mail"]
		}`))
	}))
	t.Cleanup(server.Close)

	client, err := sourcetest.NewClient(server)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	source := securitytrails.New()
	source.UseKeys("test-key")

	cfg := &sources.Configuration{
		Client:    client,
		Domain:    "example.com",
		Extractor: sources.CompileExtractor("example.com"),
	}

	results := source.Run(context.Background(), cfg, "example.com")

	subdomains, errs := sourcetest.Drain(results)

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	want := []string{"api.example.com", "www.example.com", "mail.example.com"}

	if !reflect.DeepEqual(subdomains, want) {
		t.Fatalf("subdomains = %v, want %v", subdomains, want)
	}
}

func TestSourceRunMalformedResponse(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"subdomain_count": "not-a-number"}`))
	}))
	t.Cleanup(server.Close)

	client, err := sourcetest.NewClient(server)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	source := securitytrails.New()
	source.UseKeys("test-key")

	cfg := &sources.Configuration{
		Client:    client,
		Domain:    "example.com",
		Extractor: sources.CompileExtractor("example.com"),
	}

	results := source.Run(context.Background(), cfg, "example.com")

	subdomains, errs := sourcetest.Drain(results)

	if len(subdomains) > 0 {
		t.Fatalf("unexpected subdomains: %v", subdomains)
	}

	if len(errs) != 1 {
		t.Fatalf("errors = %v, want exactly one parse error", errs)
	}
}
