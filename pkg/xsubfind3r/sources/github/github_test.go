package github_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	hqgohttp "github.com/hueristiq/hq-lib-http-go"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/github"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestRunMapsSubdomains(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/search/code" {
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write([]byte(`{"items":[{"html_url":"https://github.com/org/repo/blob/main/file","text_matches":[{"fragment":"api.example.com"}]}]}`))

			return
		}

		_, _ = writer.Write([]byte("api.example.com"))
	}))
	defer server.Close()

	target, _ := url.Parse(server.URL)
	transport := roundTripFunc(func(request *http.Request) (*http.Response, error) {
		clone := request.Clone(request.Context())
		clone.URL.Scheme = target.Scheme
		clone.URL.Host = target.Host

		return http.DefaultTransport.RoundTrip(clone)
	})

	client, err := hqgohttp.New(hqgohttp.WithHTTPClient(&http.Client{Transport: transport}))
	if err != nil {
		t.Fatalf("New HTTP client: %v", err)
	}

	source := github.New()
	source.UseKeys("secret")

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

	if len(values) != 2 || values[0] != "api.example.com" || values[1] != "api.example.com" {
		t.Fatalf("values = %v; want raw content and text match", values)
	}
}

func TestRunRequiresKeys(t *testing.T) {
	t.Parallel()

	source := github.New()

	result := <-source.Run(context.Background(), &sources.Configuration{}, "example.com")
	if result.Error == nil {
		t.Fatal("Run accepted no API keys")
	}
}
