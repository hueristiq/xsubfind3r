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

	hqgohttp "github.com/hueristiq/hq-go-http"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

// getSubdomainsResponse represents the expected JSON structure returned by the Bevigil API.
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
			Headers: []hqgohttp.Header{
				hqgohttp.NewSetHeader("X-Access-Token", key),
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

// Name returns the unique identifier for the data source.
// This identifier is used for logging, debugging, and associating results with the correct data source.
//
// Returns:
//   - name (string): The unique identifier for the data source.
func (source *Source) Name() (name string) {
	return sources.BEVIGIL
}
