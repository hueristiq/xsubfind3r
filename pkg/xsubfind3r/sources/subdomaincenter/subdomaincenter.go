// Package subdomaincenter provides an implementation of the sources.Source interface
// for interacting with the Subdomain Center API.
//
// The Subdomain Center API offers subdomain discovery for a given domain.
// This package defines a Source type that implements the Run and Name methods as specified
// by the sources.Source interface. The Run method sends a query to the Subdomain Center API,
// processes the JSON response, and streams discovered subdomains or errors via a channel.
package subdomaincenter

import (
	"encoding/json"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	hqgohttp "go.source.hueristiq.com/http"
)

// Source represents the Subdomain Center data source implementation.
// It implements the sources.Source interface, providing functionality
// for retrieving subdomains from the Subdomain Center API.
type Source struct{}

// Run initiates the process of retrieving subdomain information from the Subdomain Center API for a given domain.
//
// It constructs an HTTP GET request to the Subdomain Center API endpoint, decodes the JSON response,
// and streams each discovered subdomain as a sources.Result via a channel.
//
// Parameters:
//   - domain (string): The target domain for which to retrieve subdomains.
//   - _ (*sources.Configuration): The configuration settings (not used in this implementation).
//
// Returns:
//   - (<-chan sources.Result): A channel that asynchronously emits sources.Result values.
//     Each result is either a discovered subdomain or an error encountered during the operation.
//
// The function executes the following steps:
//  1. Constructs the API request URL and configures the required query parameters (with the target domain).
//  2. Sends an HTTP GET request using the hqgohttp package.
//  3. Decodes the JSON response into a slice of strings representing discovered subdomains.
//  4. Streams each discovered subdomain as a sources.Result of type ResultSubdomain.
//  5. Closes the results channel upon completion.
func (source *Source) Run(domain string, _ *sources.Configuration) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		getSubdomainsReqURL := "https://api.subdomain.center"
		getSubdomainsReqCFG := &hqgohttp.RequestConfiguration{
			Params: map[string]string{
				"domain": domain,
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

		var getSubdomainsResData []string

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

		for _, subdomain := range getSubdomainsResData {
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

// Name returns the unique identifier for the Subdomain Center data source.
// This identifier is used for logging, debugging, and to associate results
// with the correct data source.
//
// Returns:
//   - name (string): The constant sources.SUBDOMAINCENTER representing the Subdomain Center source.
func (source *Source) Name() (name string) {
	return sources.SUBDOMAINCENTER
}
