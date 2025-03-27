// Package shodan provides an implementation of the sources.Source interface
// for interacting with the Shodan API.
//
// The Shodan API offers DNS information for a given domain, including discovered
// subdomains. This package defines a Source type that implements the Run and Name methods as specified
// by the sources.Source interface. The Run method sends a query to the Shodan API, processes the JSON response,
// and streams discovered subdomains or errors via a channel.
package shodan

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	hqgohttp "go.source.hueristiq.com/http"
)

// getDNSResponse represents the structure of the JSON response returned by the Shodan API.
//
// It contains the following fields:
//   - Domain: A string representing the target domain for which the DNS query was performed.
//   - Subdomains: A slice of strings representing the discovered subdomains.
//   - Result: An integer indicating the status of the DNS query.
//   - Error: A string containing error information if the request encountered an issue.
type getDNSResponse struct {
	Domain     string   `json:"domain"`
	Subdomains []string `json:"subdomains"`
	Result     int      `json:"result"`
	Error      string   `json:"error"`
}

// Source represents the Shodan data source implementation.
// It implements the sources.Source interface, providing functionality
// for retrieving subdomains from the Shodan API.
type Source struct{}

// Run initiates the process of retrieving subdomain information from the Shodan API for a given domain.
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

		key, err := cfg.Keys.Shodan.PickRandom()
		if key == "" || err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		getDNSReqURL := fmt.Sprintf("https://api.shodan.io/dns/domain/%s", domain)
		getDNSReqCFG := &hqgohttp.RequestConfiguration{
			Params: map[string]string{
				"key": key,
			},
		}

		getDNSRes, err := hqgohttp.Get(getDNSReqURL, getDNSReqCFG)
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var getDNSResData getDNSResponse

		if err = json.NewDecoder(getDNSRes.Body).Decode(&getDNSResData); err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			getDNSRes.Body.Close()

			return
		}

		getDNSRes.Body.Close()

		for _, subdomain := range getDNSResData.Subdomains {
			result := sources.Result{
				Type:   sources.ResultSubdomain,
				Source: source.Name(),
				Value:  fmt.Sprintf("%s.%s", subdomain, domain),
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
	return sources.SHODAN
}
