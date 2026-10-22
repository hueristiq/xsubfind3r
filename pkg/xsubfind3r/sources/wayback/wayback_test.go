//nolint:testpackage // white-box tests replace the package-level rate limiter.
package wayback

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"regexp"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	hqgohttp "github.com/hueristiq/hq-lib-http-go"
	hqgolimiter "github.com/hueristiq/hq-lib-limiter-go"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

const exampleDomain = "example.com"

func TestMain(testingMain *testing.M) {
	limiter = hqgolimiter.New(hqgolimiter.WithRate(1_000_000, time.Second))

	os.Exit(testingMain.Run())
}

func TestPaginationCap(t *testing.T) {
	t.Parallel()

	var requests atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)

		_, _ = fmt.Fprint(w, `[["timestamp","original"],["1","s.example.com"]]`)
	}))
	defer server.Close()

	_, warnings, errs := runSource(t, server)

	if requests.Load() != sources.MaxPages {
		t.Fatalf("requests = %d, want %d", requests.Load(), sources.MaxPages)
	}

	if warnings != 1 {
		t.Fatalf("warnings = %d, want 1", warnings)
	}

	if len(errs) != 0 {
		t.Fatalf("errors = %v", errs)
	}
}

func TestEmptyPageStops(t *testing.T) {
	t.Parallel()

	var requests atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests.Add(1)

		_, _ = fmt.Fprint(w, `[]`)
	}))
	defer server.Close()

	_, warnings, errs := runSource(t, server)

	if requests.Load() != 1 {
		t.Fatalf("requests = %d, want 1", requests.Load())
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusForbidden)
	}))
	defer server.Close()

	values, warnings, errs := runSource(t, server)

	if len(values) != 0 {
		t.Fatalf("values = %v, want none", values)
	}

	if warnings != 0 {
		t.Fatalf("warnings = %d, want 0", warnings)
	}

	if len(errs) != 1 || !strings.Contains(errs[0].Error(), "unexpected status") {
		t.Fatalf("errors = %v", errs)
	}
}

func runSource(t *testing.T, server *httptest.Server) (values []string, warnings int, errs []error) {
	t.Helper()

	target, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse server URL: %v", err)
	}

	client, err := hqgohttp.NewClient(
		hqgohttp.WithHTTPClient(&http.Client{Transport: &rewriteTransport{target: target}}),
		hqgohttp.WithMaxRetries(0),
		hqgohttp.WithTimeout(5*time.Second),
	)
	if err != nil {
		t.Fatalf("NewClient: %v", err)
	}

	configuration := &sources.Configuration{
		Client:    client,
		Extractor: extractor(),
		Validator: func(target string) (subdomain string, valid bool) {
			return target, true
		},
	}

	source := New()

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

func extractor() *regexp.Regexp {
	return regexp.MustCompile(`(?i)(?:^|[^a-z0-9_.-])(((?:[a-z0-9_](?:[a-z0-9_-]*[a-z0-9_])?\.)+)example\.com)`)
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
