// Package sourcetest provides helpers for testing the sources packages:
// an hq-lib-http client that routes every request to an httptest server
// regardless of the URL the source used, plus a results drainer.
package sourcetest

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"

	hqgohttp "github.com/hueristiq/hq-lib-http-go"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

// NewClient returns an hq-lib-http client that routes every request to
// server, so source tests can serve fixtures for the hardcoded API URLs
// without network access. Retries are disabled.
func NewClient(server *httptest.Server) (client *hqgohttp.Client, err error) {
	transport := &routingTransport{target: server.URL}

	client, err = hqgohttp.NewClient(
		hqgohttp.WithHTTPClient(&http.Client{Transport: transport}),
		hqgohttp.WithMaxRetries(0),
	)
	if err != nil {
		return
	}

	return
}

// Drain consumes results until the channel closes, returning the subdomain
// values and source errors in arrival order.
func Drain(results <-chan sources.Result) (subdomains []string, errs []error) {
	for result := range results {
		switch result.Type {
		case sources.ResultSubdomain:
			subdomains = append(subdomains, result.Value)
		case sources.ResultError:
			errs = append(errs, result.Error)
		}
	}

	return
}

type routingTransport struct {
	target string
}

// RoundTrip routes the request to the target server, discarding the
// request's own URL host.
func (rt *routingTransport) RoundTrip(req *http.Request) (res *http.Response, err error) {
	target, err := url.Parse(rt.target)
	if err != nil {
		return
	}

	req = req.Clone(req.Context())

	req.URL.Scheme = target.Scheme
	req.URL.Host = target.Host
	req.Host = target.Host

	res, err = http.DefaultTransport.RoundTrip(req)
	if err != nil {
		err = fmt.Errorf("failed to route request to fixture server: %w", err)

		return
	}

	return
}
