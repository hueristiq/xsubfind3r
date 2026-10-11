package certificatedetails_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/certificatedetails"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/sourcetest"
)

const testDomain = "example.com"

func TestSourceRun(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/"+testDomain {
			t.Errorf("unexpected path %q", r.URL.Path)
		}

		_, _ = w.Write([]byte(`
<html>
	<div>api.example.com</div>
	<div>notexample.com</div>
	<div>example.com.evil.com</div>
	<div>www.example.com</div>
</html>`))
	}))
	t.Cleanup(server.Close)

	client, err := sourcetest.NewClient(server)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	source := certificatedetails.New()

	cfg := &sources.Configuration{
		Client:    client,
		Domain:    testDomain,
		Extractor: sources.CompileExtractor(testDomain),
	}

	results := source.Run(context.Background(), cfg, testDomain)

	subdomains, errs := sourcetest.Drain(results)

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	want := []string{"api.example.com", "www.example.com"}

	if !reflect.DeepEqual(subdomains, want) {
		t.Fatalf("subdomains = %v, want %v", subdomains, want)
	}
}

func TestSourceRunNotFound(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(server.Close)

	client, err := sourcetest.NewClient(server)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	source := certificatedetails.New()

	cfg := &sources.Configuration{
		Client:    client,
		Domain:    testDomain,
		Extractor: sources.CompileExtractor(testDomain),
	}

	results := source.Run(context.Background(), cfg, testDomain)

	subdomains, errs := sourcetest.Drain(results)

	if len(errs) > 0 {
		t.Fatalf("unexpected errors: %v", errs)
	}

	if len(subdomains) > 0 {
		t.Fatalf("unexpected subdomains: %v", subdomains)
	}
}

func TestSourceRunRequestFailure(t *testing.T) {
	t.Parallel()

	// A server that is closed before the request simulates a network
	// failure: the client must surface an error result instead of
	// dereferencing the nil response.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	server.Close()

	client, err := sourcetest.NewClient(server)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	source := certificatedetails.New()

	cfg := &sources.Configuration{
		Client:    client,
		Domain:    testDomain,
		Extractor: sources.CompileExtractor(testDomain),
	}

	results := source.Run(context.Background(), cfg, testDomain)

	subdomains, errs := sourcetest.Drain(results)

	if len(subdomains) > 0 {
		t.Fatalf("unexpected subdomains: %v", subdomains)
	}

	if len(errs) != 1 {
		t.Fatalf("errors = %v, want exactly one request error", errs)
	}
}
