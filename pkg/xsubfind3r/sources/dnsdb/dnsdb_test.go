//nolint:testpackage // white-box tests cover unexported offset parsing.
package dnsdb

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	hqgohttp "github.com/hueristiq/hq-lib-http-go"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

const exampleDomain = "example.com"

func TestParseOffsetMax(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		raw  string
		want uint64
	}{
		{name: "number", raw: `100`, want: 100},
		{name: "numeric string", raw: `"200"`, want: 200},
		{name: "not available", raw: `"n/a"`, want: 0},
		{name: "uppercase not available", raw: `"N/A"`, want: 0},
		{name: "null", raw: jsonNull, want: 0},
		{name: "empty string", raw: `""`, want: 0},
		{name: "garbage", raw: `"abc"`, want: 0},
		{name: "object", raw: `{}`, want: 0},
		{name: "missing", raw: ``, want: 0},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if got := parseOffsetMax(json.RawMessage(test.raw)); got != test.want {
				t.Fatalf("parseOffsetMax(%s) = %d, want %d", test.raw, got, test.want)
			}
		})
	}
}

func TestRunProcessesFinalNDJSONLine(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/rate_limit"):
			_, _ = fmt.Fprint(w, `{"rate":{"offset_max":"n/a"}}`)
		case strings.Contains(r.URL.Path, "/lookup/"):
			_, _ = fmt.Fprint(w, "{\"cond\":\"begin\",\"obj\":{\"rrname\":\"a.example.com\"}}\n{\"cond\":\"succeeded\",\"obj\":{\"rrname\":\"\"}}")
		default:
			http.NotFound(w, r)
		}
	}))
	defer server.Close()

	source := New()
	source.UseKeys("test-key")

	values, errs := collectResults(t, source.Run(t.Context(), testConfiguration(t, server), exampleDomain))

	if len(errs) != 0 {
		t.Fatalf("errors = %v", errs)
	}

	if len(values) != 1 || values[0] != "a.example.com" {
		t.Fatalf("values = %v, want [a.example.com]", values)
	}
}

func TestRunRateLimitStatusError(t *testing.T) {
	t.Parallel()

	var lookups atomic.Int64

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/lookup/") {
			lookups.Add(1)
		}

		http.Error(w, "boom", http.StatusForbidden)
	}))
	defer server.Close()

	source := New()
	source.UseKeys("test-key")

	values, errs := collectResults(t, source.Run(t.Context(), testConfiguration(t, server), exampleDomain))

	if len(values) != 0 {
		t.Fatalf("values = %v, want none", values)
	}

	if len(errs) != 1 || !strings.Contains(errs[0].Error(), "failed to get rate limit") {
		t.Fatalf("errors = %v", errs)
	}

	if lookups.Load() != 0 {
		t.Fatalf("lookups = %d, want 0", lookups.Load())
	}
}

func TestRunLookupStatusError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/rate_limit") {
			_, _ = fmt.Fprint(w, `{"rate":{"offset_max":"n/a"}}`)

			return
		}

		http.Error(w, "boom", http.StatusForbidden)
	}))
	defer server.Close()

	source := New()
	source.UseKeys("test-key")

	values, errs := collectResults(t, source.Run(t.Context(), testConfiguration(t, server), exampleDomain))

	if len(values) != 0 {
		t.Fatalf("values = %v, want none", values)
	}

	if len(errs) != 1 || !strings.Contains(errs[0].Error(), "unexpected status") {
		t.Fatalf("errors = %v", errs)
	}
}

func collectResults(t *testing.T, results <-chan sources.Result) (values []string, errs []error) {
	t.Helper()

	for result := range results {
		switch result.Type {
		case sources.ResultSubdomain:
			values = append(values, result.Value)
		case sources.ResultError:
			errs = append(errs, result.Error)
		case sources.ResultWarning:
		}
	}

	return
}

func testConfiguration(t *testing.T, server *httptest.Server) *sources.Configuration {
	t.Helper()

	target, err := url.Parse(server.URL)
	if err != nil {
		t.Fatalf("parse server URL: %v", err)
	}

	client := hqgohttp.New(
		hqgohttp.WithHTTPClient(&http.Client{Transport: &rewriteTransport{target: target}}),
		hqgohttp.WithTimeout(5*time.Second),
	)

	return &sources.Configuration{
		Client: client,
		Validator: func(target string) (subdomain string, valid bool) {
			return target, true
		},
	}
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
