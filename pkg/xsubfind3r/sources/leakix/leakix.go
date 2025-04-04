// Package leakix provides an implementation of the sources.Source interface
// for interacting with the LeakIX API.
//
// The LeakIX API offers subdomain discovery by returning detailed information about
// discovered subdomains, including distinct IP counts and the last seen timestamp.
// This package defines a Source type that implements the Run and Name methods as specified
// by the sources.Source interface. The Run method sends a query to the LeakIX API, processes
// the JSON response, and streams discovered subdomains or errors via a channel.
package leakix

import (
	"encoding/json"
	"time"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	hqgohttp "go.source.hueristiq.com/http"
	"go.source.hueristiq.com/http/header"
	"go.source.hueristiq.com/http/mime"
)

// getSubdomainsResponse represents the structure of the JSON response returned by the LeakIX API.
//
// It contains the following fields:
//   - Subdomain: A string representing the discovered subdomain.
//   - DistinctIps: An integer indicating the number of distinct IPs associated with the subdomain.
//   - LastSeen: A timestamp indicating when the subdomain was last seen.
type getSubdomainsResponse struct {
	Subdomain   string    `json:"subdomain"`
	DistinctIps int       `json:"distinct_ips"`
	LastSeen    time.Time `json:"last_seen"`
}

// Source represents the LeakIX data source implementation.
// It implements the sources.Source interface, providing functionality
// for retrieving subdomains from the LeakIX API.
type Source struct{}

// Run initiates the process of retrieving subdomain information from the LeakIX API for a given domain.
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

		key, err := cfg.Keys.LeakIX.PickRandom()
		if key == "" || err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		getSubdomainsReqURL := "https://leakix.net/api/subdomains/" + domain
		getSubdomainsReqCFG := &hqgohttp.RequestConfiguration{
			Headers: map[string]string{
				header.Accept.String(): mime.JSON.String(),
				"api-key":              key,
			},
		}

		getSubdomainsRes, err := hqgohttp.Get(getSubdomainsReqURL, getSubdomainsReqCFG)
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var getSubdomainsResData []getSubdomainsResponse

		if err = json.NewDecoder(getSubdomainsRes.Body).Decode(&getSubdomainsResData); err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			getSubdomainsRes.Body.Close()

			return
		}

		getSubdomainsRes.Body.Close()

		for _, record := range getSubdomainsResData {
			result := sources.Result{
				Type:   sources.ResultSubdomain,
				Source: source.Name(),
				Value:  record.Subdomain,
			}

			results <- result
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
	return sources.LEAKIX
}
