//nolint:testpackage // sourcetests is a test-only package, not a mirror of a source package.
package sourcetests

import (
	"net/http"
	"strings"
	"testing"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/bevigil"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/builtwith"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/c99"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/censys"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/certspotter"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/chaos"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/dnsdb"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/dnsdumpster"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/fofa"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/fullhunt"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/gitlab"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/leakix"
)

//nolint:goconst // Fixture hostnames intentionally repeat across the table.
func TestKeyedSources(t *testing.T) {
	t.Parallel()

	tests := []sourceTestCase{
		{
			name:   "bevigil",
			source: bevigil.New(),
			keys:   []string{"test-key"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got := r.Header.Get("X-Access-Token"); got != "test-key" {
					t.Errorf("X-Access-Token = %q, want %q", got, "test-key")
				}

				_, _ = w.Write([]byte(`{"domain":"example.com","subdomains":["api.example.com","www.example.com"]}`))
			},
			want: []string{"api.example.com", "www.example.com"},
		},
		{
			name:   "builtwith",
			source: builtwith.New(),
			keys:   []string{"test-key"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Query().Get("LOOKUP"); got != testDomain {
					t.Errorf("LOOKUP = %q, want %q", got, testDomain)
				}

				_, _ = w.Write([]byte(`{
					"Results": [{
						"Result": {
							"Paths": [
								{"Domain": "example.com", "SubDomain": "api"},
								{"Domain": "example.com", "SubDomain": "www"}
							]
						}
					}]
				}`))
			},
			want: []string{"api.example.com", "www.example.com"},
		},
		{
			name:   "c99",
			source: c99.New(),
			keys:   []string{"test-key"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Query().Get("key"); got != "test-key" {
					t.Errorf("key = %q, want %q", got, "test-key")
				}

				_, _ = w.Write([]byte(`{"success":true,"subdomains":[{"subdomain":"api.example.com"},{"subdomain":".www.example.com"}]}`))
			},
			want: []string{"api.example.com"},
		},
		{
			name:   "censys",
			source: censys.New(),
			keys:   []string{"test-id:test-secret"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Header.Get("Authorization") == "" {
					t.Error("missing Authorization header")
				}

				_, _ = w.Write([]byte(`{
					"code": 200,
					"status": "OK",
					"result": {
						"hits": [{"names": ["api.example.com", "www.example.com"]}],
						"links": {"next": ""}
					}
				}`))
			},
			want: []string{"api.example.com", "www.example.com"},
		},
		{
			name:   "certspotter",
			source: certspotter.New(),
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Query().Get("after") != "" {
					_, _ = w.Write([]byte(`[]`))

					return
				}

				_, _ = w.Write([]byte(`[{"id":"1","dns_names":["api.example.com","notexample.com"]}]`))
			},
			want: []string{"api.example.com"},
		},
		{
			name:   "chaos",
			source: chaos.New(),
			keys:   []string{"test-key"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got := r.Header.Get("Authorization"); got != "test-key" {
					t.Errorf("Authorization = %q, want %q", got, "test-key")
				}

				_, _ = w.Write([]byte(`{"domain":"example.com","subdomains":["api","www"],"count":2}`))
			},
			want: []string{"api.example.com", "www.example.com"},
		},
		{
			name:   "dnsdb",
			source: dnsdb.New(),
			keys:   []string{"test-key"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch {
				case strings.HasSuffix(r.URL.Path, "/rate_limit"):
					_, _ = w.Write([]byte(`{"rate":{"offset_max":0}}`))
				default:
					_, _ = w.Write([]byte(
						`{"cond":"ongoing","obj":{"rrname":"api.example.com."}}` + "\n" +
							`{"cond":"succeeded","obj":{}}` + "\n",
					))
				}
			},
			want: []string{"api.example.com"},
		},
		{
			name:   "dnsdumpster",
			source: dnsdumpster.New(),
			keys:   []string{"test-key"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got := r.Header.Get("X-Api-Key"); got != "test-key" {
					t.Errorf("X-API-Key = %q, want %q", got, "test-key")
				}

				_, _ = w.Write([]byte(`{"a":[{"host":"api.example.com"}],"ns":[{"host":"ns1.example.com"}]}`))
			},
			want: []string{"api.example.com", "ns1.example.com"},
		},
		{
			name:   "fofa",
			source: fofa.New(),
			keys:   []string{"user@example.com", "test-key"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Query().Get("email"); got != "user@example.com" {
					t.Errorf("email = %q, want %q", got, "user@example.com")
				}

				_, _ = w.Write([]byte(`{"error":false,"results":["api.example.com","https://www.example.com/x"]}`))
			},
			want: []string{"api.example.com", "www.example.com"},
		},
		{
			name:   "fullhunt",
			source: fullhunt.New(),
			keys:   []string{"test-key"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got := r.Header.Get("X-Api-Key"); got != "test-key" {
					t.Errorf("X-API-KEY = %q, want %q", got, "test-key")
				}

				_, _ = w.Write([]byte(`{"hosts":["api.example.com","www.example.com"]}`))
			},
			want: []string{"api.example.com", "www.example.com"},
		},
		{
			name:   "gitlab",
			source: gitlab.New(),
			keys:   []string{"test-key"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch {
				case strings.HasPrefix(r.URL.Path, "/api/v4/projects/"):
					if !strings.Contains(r.URL.EscapedPath(), "a%20file") {
						t.Errorf("path = %q, want the space path-escaped", r.URL.Path)
					}

					if got := r.URL.Query().Get("ref"); got != "feat/x" {
						t.Errorf("ref = %q, want %q", got, "feat/x")
					}

					_, _ = w.Write([]byte("api.example.com\nnotexample.com\n"))
				default:
					_, _ = w.Write([]byte(`[{"data":"x","project_id":1,"path":"dir/a file.txt","ref":"feat/x"}]`))
				}
			},
			want: []string{"api.example.com"},
		},
		{
			name:   "leakix",
			source: leakix.New(),
			keys:   []string{"test-key"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got := r.Header.Get("Api-Key"); got != "test-key" {
					t.Errorf("api-key = %q, want %q", got, "test-key")
				}

				_, _ = w.Write([]byte(`[{"subdomain":"api.example.com"},{"subdomain":"www.example.com"}]`))
			},
			want: []string{"api.example.com", "www.example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runSourceTest(t, &tt)
		})
	}
}
