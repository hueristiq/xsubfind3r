//nolint:testpackage // sourcetests is a test-only package, not a mirror of a source package.
package sourcetests

import (
	"net/http"
	"strings"
	"testing"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/intelx"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/leakradar"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/merklemap"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/netlas"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/onyphe"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/quake"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/shodan"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/urlscan"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/virustotal"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/whoisxmlapi"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/zoomeye"
)

//nolint:goconst // Fixture hostnames intentionally repeat across the table.
func TestKeyedSourcesMore(t *testing.T) {
	t.Parallel()

	tests := []sourceTestCase{
		{
			name:   "intelx",
			source: intelx.New(),
			keys:   []string{"intelx.example.com:test-key"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch {
				case strings.HasSuffix(r.URL.Path, "/phonebook/search"):
					_, _ = w.Write([]byte(`{"id":"search-id","status":0}`))
				default:
					_, _ = w.Write([]byte(`{"selectors":[{"selectorvalue":"api.example.com"}],"status":2}`))
				}
			},
			want: []string{"api.example.com"},
		},
		{
			name:   "intelx malformed key",
			source: intelx.New(),
			keys:   []string{"not-a-host-key-pair"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				t.Error("no request expected for a malformed key")
			},
			want:     nil,
			wantErrs: 1,
		},
		{
			name:   "leakradar",
			source: leakradar.New(),
			keys:   []string{"test-key"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got := r.Header.Get("Authorization"); got != "Bearer test-key" {
					t.Errorf("Authorization = %q, want %q", got, "Bearer test-key")
				}

				_, _ = w.Write([]byte(`{"items":[{"url":"https://api.example.com/x"}],"total":1,"page":1,"page_size":100}`))
			},
			want: []string{"api.example.com"},
		},
		{
			name:   "merklemap",
			source: merklemap.New(),
			keys:   []string{"test-key"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got := r.Header.Get("User-Agent"); got != "" && strings.Contains(got, "Mozilla") {
					t.Errorf("User-Agent = %q, the hardcoded Chrome UA must be gone", got)
				}

				_, _ = w.Write([]byte(`{"count":1,"results":[{"hostname":"api.example.com"}]}`))
			},
			want: []string{"api.example.com"},
		},
		{
			name:   "netlas",
			source: netlas.New(),
			keys:   []string{"test-key"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch {
				case strings.HasSuffix(r.URL.Path, "/domains_count/"):
					_, _ = w.Write([]byte(`{"count":2}`))
				default:
					_, _ = w.Write([]byte(`[{"data":{"domain":"api.example.com"}},{"data":{"domain":"www.example.com"}}]`))
				}
			},
			want: []string{"api.example.com", "www.example.com"},
		},
		{
			name:   "onyphe",
			source: onyphe.New(),
			keys:   []string{"test-key"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{
					"results": [{
						"subdomains": ["api.example.com"],
						"hostname": "www.example.com",
						"forward": "1.2.3.4",
						"reverse": "4.3.2.1"
					}],
					"max_page": "1"
				}`))
			},
			want: []string{"api.example.com", "www.example.com"},
		},
		{
			name:   "quake",
			source: quake.New(),
			keys:   []string{"test-key"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("method = %s, want POST", r.Method)
				}

				_, _ = w.Write([]byte(`{
					"code": 0,
					"data": [{"service": {"http": {"host": "api.example.com"}}}],
					"meta": {"pagination": {"total": 1}}
				}`))
			},
			want: []string{"api.example.com"},
		},
		{
			name:   "shodan",
			source: shodan.New(),
			keys:   []string{"test-key"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Query().Get("key"); got != "test-key" {
					t.Errorf("key = %q, want %q", got, "test-key")
				}

				_, _ = w.Write([]byte(`{"domain":"example.com","subdomains":["api","www"]}`))
			},
			want: []string{"api.example.com", "www.example.com"},
		},
		{
			name:   "urlscan",
			source: urlscan.New(),
			keys:   []string{"test-key"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got := r.Header.Get("Api-Key"); got != "test-key" {
					t.Errorf("API-Key = %q, want %q", got, "test-key")
				}

				_, _ = w.Write([]byte(`{"results":[{"page":{"domain":"api.example.com"}}],"status":200,"total":1,"has_more":false}`))
			},
			want: []string{"api.example.com"},
		},
		{
			name:   "virustotal",
			source: virustotal.New(),
			keys:   []string{"test-key"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got := r.Header.Get("X-Apikey"); got != "test-key" {
					t.Errorf("x-apikey = %q, want %q", got, "test-key")
				}

				_, _ = w.Write([]byte(`{"data":[{"id":"api.example.com","type":"domain"}],"meta":{"cursor":""}}`))
			},
			want: []string{"api.example.com"},
		},
		{
			name:   "whoisxmlapi",
			source: whoisxmlapi.New(),
			keys:   []string{"test-key"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if got := r.URL.Query().Get("apiKey"); got != "test-key" {
					t.Errorf("apiKey = %q, want %q", got, "test-key")
				}

				_, _ = w.Write([]byte(`{"result":{"records":[{"domain":"api.example.com"},{"domain":"www.example.com"}]}}`))
			},
			want: []string{"api.example.com", "www.example.com"},
		},
		{
			name:   "zoomeye",
			source: zoomeye.New(),
			keys:   []string{"test-key"},
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.Method != http.MethodPost {
					t.Errorf("method = %s, want POST", r.Method)
				}

				_, _ = w.Write([]byte(`{"total":1,"data":[{"domain":"api.example.com","url":"https://unrelated.example.net/x"}]}`))
			},
			want: []string{"api.example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runSourceTest(t, &tt)
		})
	}
}
