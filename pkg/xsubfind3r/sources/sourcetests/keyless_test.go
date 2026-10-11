// Package sourcetests runs table-driven decode tests for every source
// against httptest-served fixtures: responses captured in the shape each
// API returns, decoded by the real source implementation, with junk
// look-alikes that must not leak into the results.
//
//nolint:testpackage // sourcetests is a test-only package, not a mirror of a source package.
package sourcetests

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"
	"time"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/anubis"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/arquivo"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/commoncrawl"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/crtsh"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/driftnet"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/ghostarchive"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/hackertarget"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/hudsonrock"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/otx"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/rapiddns"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/riddler"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/sourcetest"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/subdomaincenter"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/threatminer"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/wayback"
)

const testDomain = "example.com"

type sourceTestCase struct {
	name     string
	source   sources.Source
	keys     []string
	handler  http.HandlerFunc
	want     []string
	wantErrs int
}

func runSourceTest(t *testing.T, tt *sourceTestCase) {
	t.Helper()

	var server *httptest.Server //nolint:staticcheck // server must exist before the handler closure references it.

	server = httptest.NewServer(tt.handler)
	t.Cleanup(server.Close)

	client, err := sourcetest.NewClient(server)
	if err != nil {
		t.Fatalf("failed to create client: %v", err)
	}

	source := tt.source

	if len(tt.keys) > 0 {
		source.UseKeys(tt.keys...)
	}

	cfg := &sources.Configuration{
		Client:    client,
		Domain:    testDomain,
		Extractor: sources.CompileExtractor(testDomain),
	}

	results := source.Run(context.Background(), cfg, testDomain)

	subdomains, errs := sourcetest.Drain(results)

	if len(errs) != tt.wantErrs {
		t.Fatalf("errors = %v, want %d", errs, tt.wantErrs)
	}

	if !reflect.DeepEqual(subdomains, tt.want) {
		t.Fatalf("subdomains = %v, want %v", subdomains, tt.want)
	}
}

//nolint:goconst // Fixture hostnames intentionally repeat across the table.
func TestKeylessSources(t *testing.T) {
	t.Parallel()

	tests := []sourceTestCase{
		{
			name:   "anubis",
			source: anubis.New(),
			handler: func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/anubis/subdomains/"+testDomain {
					t.Errorf("unexpected path %q", r.URL.Path)
				}

				_, _ = w.Write([]byte(`["api.example.com","www.example.com","notexample.com"]`))
			},
			want: []string{"api.example.com", "www.example.com", "notexample.com"},
		},
		{
			name:   "arquivo",
			source: arquivo.New(),
			handler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(
					`{"url":"https://api.example.com/index"}` + "\n" +
						`{"url":"https://notexample.com/index"}` + "\n",
				))
			},
			want: []string{"api.example.com"},
		},
		{
			name:   "commoncrawl",
			source: commoncrawl.New(),
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Path {
				case "/collinfo.json":
					_, _ = fmt.Fprintf(w, `[{"id":"CC-MAIN-%d-01","cdx-api":"https://index.commoncrawl.org/CC-MAIN-%d-01"}]`,
						time.Now().Year(), time.Now().Year())
				case "/CC-MAIN-" + strconv.Itoa(time.Now().Year()) + "-01":
					if r.URL.Query().Get("showNumPages") == "true" {
						_, _ = w.Write([]byte(`{"blocks":1,"pageSize":1,"pages":1}`))
					} else {
						_, _ = w.Write([]byte(`{"url":"https://api.example.com/index"}` + "\n" + `{"url":"https://notexample.com/index"}` + "\n"))
					}
				default:
					http.NotFound(w, r)
				}
			},
			want: []string{"api.example.com"},
		},
		{
			name:   "crtsh",
			source: crtsh.New(),
			handler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`[{"id":1,"name_value":"api.example.com\nwww.example.com\nnotexample.com"}]`))
			},
			want: []string{"api.example.com", "www.example.com"},
		},
		{
			name:   "driftnet",
			source: driftnet.New(),
			handler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{
					"enrichments": {},
					"observations": {
						"host": {
							"values": {
								"obs1": {
									"api.example.com": 0
								},
								"obs2": {
									"notexample.com": 0
								}
							}
						},
						"subject;cert": {
							"values": {
								"obs1": {
									"www.example.com": 1
								}
							}
						}
					}
				}`))
			},
			want: []string{"api.example.com", "www.example.com"},
		},
		{
			name:   "ghostarchive",
			source: ghostarchive.New(),
			handler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(
					`<html><a href="/archive/20240101000000/api.example.com">api.example.com</a>` +
						`<a href="/archive/20240101000000/notexample.com">notexample.com</a></html>`,
				))
			},
			want: []string{"api.example.com"},
		},
		{
			name:   "hackertarget",
			source: hackertarget.New(),
			handler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("api.example.com,1.2.3.4\nnotexample.com,1.2.3.4\n"))
			},
			want: []string{"api.example.com"},
		},
		{
			name:   "hudsonrock",
			source: hudsonrock.New(),
			handler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{
					"data": {
						"employees_urls": [{"url": "https://api.example.com/x"}],
						"clients_urls": [{"url": "https://www.example.com/x"}]
					}
				}`))
			},
			want: []string{"api.example.com", "www.example.com"},
		},
		{
			name:   "otx",
			source: otx.New(),
			handler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"passive_dns":[{"hostname":"api.example.com"},{"hostname":"notexample.com"}]}`))
			},
			want: []string{"api.example.com"},
		},
		{
			name:   "rapiddns",
			source: rapiddns.New(),
			handler: func(w http.ResponseWriter, r *http.Request) {
				switch r.URL.Query().Get("page") {
				case "1":
					_, _ = w.Write([]byte(`<html><td>api.example.com</td><a class="page-link" href="/subdomain/example.com?page=2">2</a></html>`))
				default:
					_, _ = w.Write([]byte(`<html><td>www.example.com</td><td>notexample.com</td></html>`))
				}
			},
			want: []string{"api.example.com", "www.example.com"},
		},
		{
			name:   "riddler",
			source: riddler.New(),
			handler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte("api.example.com\nnotexample.com\n"))
			},
			want: []string{"api.example.com"},
		},
		{
			name:   "subdomaincenter",
			source: subdomaincenter.New(),
			handler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`["api.example.com","www.example.com"]`))
			},
			want: []string{"api.example.com", "www.example.com"},
		},
		{
			name:   "threatminer",
			source: threatminer.New(),
			handler: func(w http.ResponseWriter, r *http.Request) {
				_, _ = w.Write([]byte(`{"status_code":"200","status_message":"Results found.","results":["api.example.com","www.example.com"]}`))
			},
			want: []string{"api.example.com", "www.example.com"},
		},
		{
			name:   "wayback",
			source: wayback.New(),
			handler: func(w http.ResponseWriter, r *http.Request) {
				// fl=original yields single-column CDX rows; the first
				// result is skipped by the source. Pages past the first
				// are empty so pagination terminates.
				if r.URL.Query().Get("page") != "0" {
					_, _ = w.Write([]byte(`[]`))

					return
				}

				_, _ = w.Write([]byte(`[["https://api.example.com/"],["https://www.example.com/"]]`))
			},
			want: []string{"www.example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			runSourceTest(t, &tt)
		})
	}
}
