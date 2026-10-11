package github_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"sync"
	"testing"
	"time"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/github"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/sourcetest"
)

const (
	testDomain = "example.com"
	searchPath = "/search/code"
)

func TestSourceRun(t *testing.T) {
	t.Parallel()

	var (
		mu          sync.Mutex
		rawRequests int
		server      *httptest.Server
	)

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case searchPath:
			if got := r.URL.Query().Get("q"); got != `"`+testDomain+`"` {
				t.Errorf("q = %q, want %q", got, `"`+testDomain+`"`)
			}

			w.Header().Set("Content-Type", "application/json")

			_, _ = w.Write([]byte(`{
				"total_count": 1,
				"items": [{
					"name": "subs.txt",
					"html_url": "` + server.URL + `/raw/subs.txt",
					"text_matches": [{"fragment": "found x.example.com in code"}]
				}]
			}`))
		case "/raw/subs.txt":
			mu.Lock()
			rawRequests++
			mu.Unlock()

			_, _ = w.Write([]byte("api.example.com\nnotexample.com\nexample.com.evil.com\n"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	client, err := sourcetest.NewClient(server)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	source := github.New()
	source.UseKeys("test-key")

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

	want := []string{"api.example.com", "x.example.com"}

	if !reflect.DeepEqual(subdomains, want) {
		t.Fatalf("subdomains = %v, want %v", subdomains, want)
	}

	mu.Lock()
	defer mu.Unlock()

	if rawRequests != 1 {
		t.Fatalf("raw content requests = %d, want 1", rawRequests)
	}
}

func TestSourceRunRateLimitRotatesKeys(t *testing.T) {
	t.Parallel()

	var (
		mu             sync.Mutex
		authorizations []string
		server         *httptest.Server
	)

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != searchPath {
			http.NotFound(w, r)

			return
		}

		mu.Lock()

		authorizations = append(authorizations, r.Header.Get("Authorization"))

		mu.Unlock()

		if len(authorizations) == 1 {
			w.Header().Set("X-RateLimit-Remaining", "0")
			w.Header().Set("X-Ratelimit-Reset", strconv.FormatInt(time.Now().Add(time.Hour).Unix(), 10))
			w.WriteHeader(http.StatusForbidden)

			return
		}

		_, _ = w.Write([]byte(`{"total_count": 0, "items": []}`))
	}))
	t.Cleanup(server.Close)

	client, err := sourcetest.NewClient(server)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	source := github.New()
	source.UseKeys("key-a", "key-b")

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

	if len(subdomains) != 0 {
		t.Fatalf("unexpected subdomains: %v", subdomains)
	}

	mu.Lock()
	defer mu.Unlock()

	if len(authorizations) != 2 {
		t.Fatalf("requests = %d, want 2 (one per key)", len(authorizations))
	}

	if authorizations[0] != "token key-a" {
		t.Fatalf("first request used %q, want key-a", authorizations[0])
	}

	if authorizations[1] != "token key-b" {
		t.Fatalf("second request used %q, want key-b (the exhausted key must rotate out)", authorizations[1])
	}
}

func TestSourceRunRateLimitTerminates(t *testing.T) {
	t.Parallel()

	var (
		mu       sync.Mutex
		requests int
		server   *httptest.Server
	)

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		requests++
		mu.Unlock()

		w.Header().Set("X-RateLimit-Remaining", "0")
		w.WriteHeader(http.StatusForbidden)
	}))
	t.Cleanup(server.Close)

	client, err := sourcetest.NewClient(server)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	source := github.New()
	source.UseKeys("test-key")

	cfg := &sources.Configuration{
		Client:    client,
		Domain:    testDomain,
		Extractor: sources.CompileExtractor(testDomain),
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	results := source.Run(ctx, cfg, testDomain)

	subdomains, errs := sourcetest.Drain(results)

	if len(subdomains) > 0 {
		t.Fatalf("unexpected subdomains: %v", subdomains)
	}

	if len(errs) != 1 {
		t.Fatalf("errors = %v, want exactly one rate-limit error", errs)
	}

	mu.Lock()
	defer mu.Unlock()

	if requests != 1 {
		t.Fatalf("requests = %d, want 1 (no recovery hint must be terminal, not a spin loop)", requests)
	}
}

func TestSourceRunPaginates(t *testing.T) {
	t.Parallel()

	var server *httptest.Server

	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == searchPath && r.URL.Query().Get("page") == "2":
			_, _ = w.Write([]byte(`{"total_count": 1, "items": [{"name": "s2", "html_url": "` + server.URL + `/raw2", "text_matches": []}]}`))
		case r.URL.Path == searchPath:
			w.Header().Set("Link", `<`+server.URL+searchPath+`?page=2>; rel="next"`)

			_, _ = w.Write([]byte(`{"total_count": 1, "items": [{"name": "s1", "html_url": "` + server.URL + `/raw1", "text_matches": []}]}`))
		case r.URL.Path == "/raw1":
			_, _ = w.Write([]byte("sub.page1.example.com\n"))
		case r.URL.Path == "/raw2":
			_, _ = w.Write([]byte("sub.page2.example.com\n"))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(server.Close)

	client, err := sourcetest.NewClient(server)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	source := github.New()
	source.UseKeys("test-key")

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

	want := []string{"sub.page1.example.com", "sub.page2.example.com"}

	if !reflect.DeepEqual(subdomains, want) {
		t.Fatalf("subdomains = %v, want %v", subdomains, want)
	}
}
