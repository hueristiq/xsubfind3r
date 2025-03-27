// Package wayback provides an implementation of the sources.Source interface
// for interacting with the Wayback Machine (Internet Archive) API.
//
// The Wayback Machine API (CDX Server API) offers access to historical snapshots
// of URLs, which can be used to discover subdomains by extracting URL information
// from archived pages. This package defines a Source type that implements the Run
// and Name methods as specified by the sources.Source interface. The Run method sends
// paginated queries to the Wayback Machine API, processes the JSON response, extracts
// subdomains using a provided regular expression, and streams discovered subdomains or
// errors via a channel.
//
// Additionally, a rate limiter is configured to control the number of requests per minute.
package wayback

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	hqgohttp "go.source.hueristiq.com/http"
	hqgolimiter "go.source.hueristiq.com/limiter"
)

// Source represents the Wayback Machine data source implementation.
// It implements the sources.Source interface, providing functionality
// for retrieving subdomains from the Wayback Machine API.
type Source struct{}

// limiter is a rate limiter configured to allow up to 40 requests per minute.
// This is used to throttle requests to the Wayback Machine API.
var limiter = hqgolimiter.New(&hqgolimiter.Configuration{
	RequestsPerMinute: 40,
})

// Run initiates the process of retrieving subdomain information from the Wayback Machine API for a given domain.
//
// Parameters:
//   - domain (string): The target domain for which to retrieve subdomains.
//   - cfg (*sources.Configuration): The configuration instance containing API keys,
//     the URL validation function, and any additional settings required by the source.
//
// Returns:
//   - (<-chan sources.Result): A channel that asynchronously emits sources.Result values.
//     Each result is either a discovered subdomain (ResultSubdomain) or an error (ResultError)
//     encountered during the operation.
func (source *Source) Run(domain string, cfg *sources.Configuration) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		for page := uint(0); ; page++ {
			limiter.Wait()

			getURLsReqURL := "https://web.archive.org/cdx/search/cdx"
			getURLsReqCFG := &hqgohttp.RequestConfiguration{
				Params: map[string]string{
					"url":      "*." + domain + "/*",
					"output":   "json",
					"collapse": "urlkey",
					"fl":       "original",
					"pageSize": "100",
					"page":     fmt.Sprintf("%d", page),
				},
			}

			getURLsRes, err := hqgohttp.Get(getURLsReqURL, getURLsReqCFG)
			if err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: source.Name(),
					Error:  err,
				}

				results <- result

				return
			}

			var getURLsResData [][]string

			if err = json.NewDecoder(getURLsRes.Body).Decode(&getURLsResData); err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: source.Name(),
					Error:  err,
				}

				results <- result

				getURLsRes.Body.Close()

				return
			}

			getURLsRes.Body.Close()

			// check if there's results, wayback's pagination response
			// is not always correct
			if len(getURLsResData) == 0 {
				break
			}

			// Slicing as [1:] to skip first result by default
			for _, entry := range getURLsResData[1:] {
				match := cfg.Extractor.FindAllString(entry[0], -1)

				for _, subdomain := range match {
					result := sources.Result{
						Type:   sources.ResultSubdomain,
						Source: source.Name(),
						Value:  subdomain,
					}

					results <- result
				}
			}
		}
	}()

	return results
}

// Name returns the unique identifier for the data source.
// This identifier is used for logging, debugging, and associating results with the correct data source.
//
// Returns:
//   - name (string): The unique identifier for the data source.
func (source *Source) Name() (name string) {
	return sources.WAYBACK
}
