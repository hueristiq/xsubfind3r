// Package chaos provides an implementation of the sources.Source interface
// for interacting with the Chaos API.
//
// The Chaos API offers subdomain discovery by querying ProjectDiscovery's DNS API.
// This package defines a Source type that implements the Run and Name methods as
// specified by the sources.Source interface. The Run method sends a query to the Chaos API,
// processes the JSON response, and streams discovered subdomains or errors via a channel.
package chaos

import (
	"encoding/json"
	"fmt"

	hqgohttp "github.com/hueristiq/hq-go-http"
	"github.com/hueristiq/hq-go-http/header"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

// getSubdomainsResponse represents the structure of the JSON response returned by the Chaos API.
//
// It contains the following fields:
//   - Domain: The primary domain for which subdomains are being queried.
//   - Subdomains: A slice of strings representing the discovered subdomains.
//   - Count: The total number of discovered subdomains.
type getSubdomainsResponse struct {
	Domain     string   `json:"domain"`
	Subdomains []string `json:"subdomains"`
	Count      int      `json:"count"`
}

// Source represents the Chaos API data source implementation.
// It implements the sources.Source interface, providing functionality
// for retrieving subdomains from the Chaos API.
type Source struct{}

// Run initiates the process of retrieving subdomain information from the Chaos API for a given domain.
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

		key, err := cfg.Keys.Chaos.PickRandom()
		if key == "" || err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		getSubdomainsReqURL := fmt.Sprintf(
			"https://dns.projectdiscovery.io/dns/%s/subdomains",
			domain,
		)
		getSubdomainsReqCFG := &hqgohttp.RequestConfiguration{
			Headers: []hqgohttp.Header{
				hqgohttp.NewHeader(header.Authorization.String(), key, hqgohttp.HeaderModeSet),
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

		var getSubdomainsResData getSubdomainsResponse

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

		for _, subdomain := range getSubdomainsResData.Subdomains {
			result := sources.Result{
				Type:   sources.ResultSubdomain,
				Source: source.Name(),
				Value:  fmt.Sprintf("%s.%s", subdomain, getSubdomainsResData.Domain),
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
	return sources.CHAOS
}
