package urlscan_test

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
	"time"

	hqgohttp "github.com/hueristiq/hq-lib-http-go"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/urlscan"
)

const exampleDomain = "example.com"

func TestPaginationCap(t *testing.T) {
	t.Parallel()

	var requests atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		number := requests.Add(1)

		_, _ = fmt.Fprintf(w, `{"results":[{"page":{"domain":"s%d.example.com"},"sort":["%d"]}],"has_more":true}`, number, number)
	}))
	defer server.Close()

	values, warnings, errs := runSource(t, server)

	if requests.Load() != sources.MaxPages {
		t.Fatalf("requests = %d, want %d", requests.Load(), sources.MaxPages)
	}

	if warnings != 1 {
		t.Fatalf("warnings = %d, want 1", warnings)
	}

	if len(errs) != 0 {
		t.Fatalf("errors = %v", errs)
	}

	if len(values) != sources.MaxPages {
		t.Fatalf("values = %d, want %d", len(values), sources.MaxPages)
	}
}

func TestRepeatedCursorStops(t *testing.T) {
	t.Parallel()

	var requests atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)

		_, _ = fmt.Fprint(w, `{"results":[{"page":{"domain":"same.example.com"},"sort":["same"]}],"has_more":true}`)
	}))
	defer server.Close()

	_, warnings, errs := runSource(t, server)

	if requests.Load() != 2 {
		t.Fatalf("requests = %d, want 2", requests.Load())
	}

	if warnings != 0 {
		t.Fatalf("warnings = %d, want 0", warnings)
	}

	if len(errs) != 0 {
		t.Fatalf("errors = %v", errs)
	}
}

func TestNon2xxReturnsError(t *testing.T) {
	t.Parallel()

	var requests atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)

		http.Error(w, "boom", http.StatusForbidden)
	}))
	defer server.Close()

	values, warnings, errs := runSource(t, server)

	if requests.Load() != 1 {
		t.Fatalf("requests = %d, want 1", requests.Load())
	}

	if len(values) != 0 {
		t.Fatalf("values = %v, want none", values)
	}

	if warnings != 0 {
		t.Fatalf("warnings = %d, want 0", warnings)
	}

	if len(errs) != 1 {
		t.Fatalf("errors = %v, want 1", errs)
	}
}

func runSource(t *testing.T, server *httptest.Server) (values []string, warnings int, errs []error) {
	t.Helper()

	target, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse server URL: %v", err)
	}

	client := hqgohttp.New(
		hqgohttp.WithHTTPClient(&http.Client{Transport: &rewriteTransport{target: target}}),
		hqgohttp.WithTimeout(5*time.Second),
	)

	source := urlscan.New()
	source.UseKeys("test-key")

	configuration := &sources.Configuration{
		Client: client,
		Validator: func(target string) (subdomain string, valid bool) {
			return target, true
		},
	}

	for result := range source.Run(t.Context(), configuration, exampleDomain) {
		switch result.Type {
		case sources.ResultSubdomain:
			values = append(values, result.Value)
		case sources.ResultWarning:
			warnings++
		case sources.ResultError:
			errs = append(errs, result.Error)
		}
	}

	return
}

type rewriteTransport struct {
	target *url.URL
}

func (transport *rewriteTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	clone := request.Clone(request.Context())
	clone.URL.Scheme = transport.target.Scheme
	clone.URL.Host = transport.target.Host

	response, err := http.DefaultTransport.RoundTrip(clone)
	if err != nil {
		return nil, fmt.Errorf("round trip: %w", err)
	}

	return response, nil
}
