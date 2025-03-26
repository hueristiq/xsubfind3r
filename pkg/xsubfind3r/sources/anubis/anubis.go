// Package anubis provides an implementation of the sources.Source interface
// for interacting with the Anubis data source.
//
// The Anubis service (hosted at jldc.me) offers an API endpoint for subdomain
// discovery by querying its database. This package defines a Source type that
// implements the Run and Name methods as specified by the sources.Source interface.
// The Run method retrieves subdomains for a given domain and streams the results
// asynchronously through a channel.
package anubis

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	hqgohttp "go.source.hueristiq.com/http"
)

// Source represents the Anubis data source implementation.
// It adheres to the sources.Source interface, providing functionality
// to retrieve subdomains using the Anubis API.
type Source struct{}

// Run initiates a subdomain discovery operation for the given domain using the Anubis API.
//
// It constructs an HTTP GET request to the Anubis API endpoint, decodes the JSON response,
// and streams each discovered subdomain as a sources.Result via a channel.
//
// Parameters:
//   - domain (string): The target domain for which subdomains are to be retrieved.
//   - _ (*sources.Configuration): The configuration settings (unused in this implementation).
//
// Returns:
//   - (<-chan sources.Result): A channel that asynchronously emits sources.Result values.
//     Each result represents either a discovered subdomain or an error encountered during the operation.
//
// The function performs the following steps:
//  1. Creates a results channel to send discovered subdomains or errors.
//  2. Constructs the URL for the Anubis API endpoint using the provided domain.
//  3. Sends an HTTP GET request to the API endpoint via the hqgohttp package.
//  4. If an error occurs during the HTTP request, sends an error result through the channel.
//  5. If the response is successful, decodes the JSON response into a slice of strings.
//  6. On JSON decoding error, sends an error result and closes the response body.
//  7. After successfully decoding the subdomains, closes the response body and sends each
//     subdomain as an individual result through the channel.
//  8. Closes the channel after all results have been processed.
func (source *Source) Run(domain string, _ *sources.Configuration) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		getSubdomainsReqURL := fmt.Sprintf("https://jldc.me/anubis/subdomains/%s", domain)

		getSubdomainsRes, err := hqgohttp.Get(getSubdomainsReqURL)
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

// Name returns the unique identifier for the Anubis data source.
// This identifier is used for logging, debugging, and associating results with the correct data source.
//
// Returns:
//   - name (string): The constant sources.ANUBIS representing the Anubis source.
func (source *Source) Name() (name string) {
	return sources.ANUBIS
}
