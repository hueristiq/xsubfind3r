// Package fullhunt provides an implementation of the sources.Source interface
// for interacting with the Fullhunt API.
//
// The Fullhunt API offers subdomain discovery for a given domain.
// This package defines a Source type that implements the Run and Name methods as specified
// by the sources.Source interface. The Run method sends a query to the Fullhunt API, processes
// the JSON response, and streams discovered subdomains or errors via a channel.
package fullhunt

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	hqgohttp "go.source.hueristiq.com/http"
)

// getSubdomainsResponse represents the structure of the JSON response returned by the Fullhunt API.
//
// It contains the following fields:
//   - Hosts: A slice of strings representing the discovered subdomains.
//   - Message: A string containing an optional message returned by the API.
//   - Status: An integer indicating the status code of the API response.
type getSubdomainsResponse struct {
	Hosts   []string `json:"hosts"`
	Message string   `json:"message"`
	Status  int      `json:"status"`
}

// Source represents the Fullhunt data source implementation.
// It implements the sources.Source interface, providing functionality
// for retrieving subdomains from the Fullhunt API.
type Source struct{}

// Run initiates the process of retrieving subdomain information from the Fullhunt API for a given domain.
//
// It constructs an HTTP GET request to the Fullhunt API endpoint, decodes the JSON response,
// and streams each discovered subdomain as a sources.Result via a channel.
//
// Parameters:
//   - domain (string): The target domain for which to retrieve subdomains.
//   - cfg (*sources.Configuration): The configuration settings (which include API keys)
//     used to authenticate with the Fullhunt API.
//
// Returns:
//   - (<-chan sources.Result): A channel that asynchronously emits sources.Result values.
//     Each result is either a discovered subdomain (ResultSubdomain) or an error (ResultError)
//     encountered during the operation.
//
// The function executes the following steps:
//  1. Attempts to retrieve a random API key from the configuration's Fullhunt keys.
//  2. Constructs the API request URL for the target domain and configures the required headers for authentication.
//  3. Sends an HTTP GET request using the hqgohttp package.
//  4. Decodes the JSON response into a getSubdomainsResponse struct.
//  5. Iterates over the discovered subdomains (hosts) and streams each as a sources.Result of type ResultSubdomain.
//  6. Closes the results channel upon completion.
func (source *Source) Run(domain string, cfg *sources.Configuration) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		key, err := cfg.Keys.Fullhunt.PickRandom()
		if key == "" || err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		getSubdomainsReqURL := fmt.Sprintf("https://fullhunt.io/api/v1/domain/%s/subdomains", domain)
		getSubdomainsReqCFG := &hqgohttp.RequestConfiguration{
			Headers: map[string]string{
				"X-API-KEY": key,
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

		err = json.NewDecoder(getSubdomainsRes.Body).Decode(&getSubdomainsResData)
		if err != nil {
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

		for _, subdomain := range getSubdomainsResData.Hosts {
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

// Name returns the unique identifier for the Fullhunt data source.
// This identifier is used for logging, debugging, and to associate results
// with the correct data source.
//
// Returns:
//   - name (string): The constant sources.FULLHUNT representing the Fullhunt source.
func (source *Source) Name() (name string) {
	return sources.FULLHUNT
}
