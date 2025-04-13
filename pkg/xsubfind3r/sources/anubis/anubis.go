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

	hqgohttp "github.com/hueristiq/hq-go-http"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

// Source represents the Anubis data source implementation.
// It adheres to the sources.Source interface, providing functionality
// to retrieve subdomains using the Anubis API.
type Source struct{}

// Run initiates a subdomain discovery operation for the given domain using the Anubis API.
//
// Parameters:
//   - domain (string): The target domain for which subdomains are to be retrieved.
//   - _ (*sources.Configuration): The configuration instance containing API keys,
//     the URL validation function, and any additional settings required by the source.
//
// Returns:
//   - (<-chan sources.Result): A channel that asynchronously emits sources.Result values.
//     Each result is either a discovered subdomain (ResultSubdomain) or an error (ResultError)
//     encountered during the operation.
func (source *Source) Run(domain string, _ *sources.Configuration) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		getSubdomainsReqURL := "https://jldc.me/anubis/subdomains/" + domain

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

// Name returns the unique identifier for the data source.
// This identifier is used for logging, debugging, and associating results with the correct data source.
//
// Returns:
//   - name (string): The unique identifier for the data source.
func (source *Source) Name() (name string) {
	return sources.ANUBIS
}
