// Package bevigil provides an implementation of the sources.Source interface
// for interacting with the Bevigil OSINT API.
//
// Bevigil's API offers subdomain discovery for a given domain. This package defines
// a Source type that implements the Run and Name methods as defined by the sources.Source interface.
// The Run method fetches subdomains for a target domain using an API key from the configuration,
// decodes the JSON response, and streams each discovered subdomain as a sources.Result via a channel.
package bevigil

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	hqgohttp "go.source.hueristiq.com/http"
)

// getSubdomainsResponse represents the expected JSON structure returned by the Bevigil API.
//
// It contains the target domain and a list of discovered subdomains.
//
// Fields:
//   - Domain (string): The domain for which the subdomains were queried.
//   - Subdomains ([]string): A slice of strings containing the discovered subdomains.
type getSubdomainsResponse struct {
	Domain     string   `json:"domain"`
	Subdomains []string `json:"subdomains"`
}

// Source represents the Bevigil data source implementation.
// It implements the sources.Source interface, providing functionality
// for retrieving subdomains from the Bevigil OSINT API.
type Source struct{}

// Run initiates the subdomain discovery process for a given domain using the Bevigil API.
//
// It constructs an HTTP GET request to the Bevigil API endpoint, decodes the JSON response,
// and streams each discovered subdomain as a sources.Result via a channel.
//
// Parameters:
//   - domain (string): The target domain for which to retrieve subdomains.
//   - cfg (*sources.Configuration): The configuration settings (which include API keys) used to authenticate with the Bevigil API.
//
// Returns:
//   - (<-chan sources.Result): A channel that asynchronously emits sources.Result values.
//     Each result is either a discovered subdomain (ResultSubdomain) or an error (ResultError)
//     encountered during the operation.
//
// The function executes the following steps:
//  1. Attempts to retrieve a random API key from the configuration for the Bevigil source.
//  2. Constructs the API request URL using the target domain.
//  3. Configures the HTTP request with the necessary authentication header using the retrieved API key.
//  4. Sends the HTTP GET request via the hqgohttp package.
//  5. If the request or JSON decoding fails, sends an error result through the results channel.
//  6. Decodes the JSON response into a getSubdomainsResponse struct.
//  7. Iterates over each discovered subdomain and sends it as an individual sources.Result of type ResultSubdomain.
//  8. Closes the results channel after processing all subdomains.
func (source *Source) Run(domain string, cfg *sources.Configuration) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		key, err := cfg.Keys.Bevigil.PickRandom()
		if key == "" || err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		getSubdomainsReqURL := fmt.Sprintf("https://osint.bevigil.com/api/%s/subdomains/", domain)
		getSubdomainsReqCFG := &hqgohttp.RequestConfiguration{
			Headers: map[string]string{
				"X-Access-Token": key,
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
				Value:  subdomain,
			}

			results <- result
		}
	}()

	return results
}

// Name returns the unique identifier for the Bevigil data source.
// This identifier is used for logging, debugging, and to associate results
// with the correct data source.
//
// Returns:
//   - name (string): The constant sources.BEVIGIL representing the Bevigil source.
func (source *Source) Name() (name string) {
	return sources.BEVIGIL
}
