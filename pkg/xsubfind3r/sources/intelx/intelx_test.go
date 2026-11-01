package intelx_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	hqgohttp "github.com/hueristiq/hq-lib-http-go"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/intelx"
)

func TestRunMapsSubdomains(t *testing.T) {
	t.Parallel()

	server := httptest.NewTLSServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		writer.Header().Set("Content-Type", "application/json")

		if r.URL.Path == "/phonebook/search" {
			_, _ = writer.Write([]byte(`{"id":"search-id"}`))

			return
		}

		_, _ = writer.Write([]byte(`{"status":1,"selectors":[{"selectorvalue":"api.example.com"}]}`))
	}))
	defer server.Close()

	client, err := hqgohttp.New(hqgohttp.WithHTTPClient(server.Client()))
	if err != nil {
		t.Fatalf("New HTTP client: %v", err)
	}

	source := intelx.New()
	source.UseKeys(strings.TrimPrefix(server.URL, "https://") + ":secret")

	cfg := &sources.Configuration{
		Client:    client,
		Extractor: sources.NewSubdomainExtractor("example.com"),
		Validator: func(value string) (string, bool) { return value, true },
	}

	var values []string

	for result := range source.Run(context.Background(), cfg, "example.com") {
		if result.Error != nil {
			t.Fatalf("Run: %v", result.Error)
		}

		if result.Type != sources.ResultSubdomain {
			t.Fatalf("result type = %v; want subdomain", result.Type)
		}

		values = append(values, result.Value)
	}

	if len(values) != 1 || values[0] != "api.example.com" {
		t.Fatalf("values = %v; want api.example.com", values)
	}
}

func TestRunRequiresCredential(t *testing.T) {
	t.Parallel()

	source := intelx.New()
	results := source.Run(context.Background(), &sources.Configuration{}, "example.com")

	result, ok := <-results
	if !ok || result.Error == nil {
		t.Fatalf("missing-credential result = %+v, %t; want an error", result, ok)
	}
}
