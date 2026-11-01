package gitlab_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	hqgohttp "github.com/hueristiq/hq-lib-http-go"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/gitlab"
)

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(request *http.Request) (*http.Response, error) {
	return f(request)
}

func TestRunMapsSubdomains(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Private-Token") != "secret" {
			t.Error("missing private token")
		}

		if r.URL.Path == "/api/v4/search" {
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write([]byte(`[{"data":"","project_id":7,"path":"dir/file","ref":"main"}]`))

			return
		}

		_, _ = writer.Write([]byte("https://api.example.com/x"))
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

	source := gitlab.New()
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

	if len(values) != 1 || values[0] != "api.example.com" {
		t.Fatalf("values = %v; want api.example.com", values)
	}
}
