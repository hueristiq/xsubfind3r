// Package certspotter provides an implementation of the sources.Source interface
// for interacting with the Certspotter API.
//
// The Certspotter API offers certificate transparency search capabilities that
// allow for the discovery of subdomains by returning certificate issuance data.
// This package defines a Source type that implements the Run and Name methods as
// specified by the sources.Source interface. The Run method sends queries to the
// Certspotter API, processes the JSON responses (including handling pagination via
// the "after" parameter), and streams discovered subdomains or errors via a channel.
package certspotter

import (
	"encoding/json"
	"strings"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	hqgohttp "go.source.hueristiq.com/http"
	"go.source.hueristiq.com/http/header"
)

// getCTLogsSearchResponse represents the structure of the JSON response returned by the Certspotter API.
//
// It contains the following fields:
//   - ID: A unique identifier for the certificate record.
//   - DNSNames: A slice of strings representing the DNS names (subdomains) associated with the certificate.
type getCTLogsSearchResponse struct {
	ID       string   `json:"id"`
	DNSNames []string `json:"dns_names"`
}

// Source represents the Certspotter data source implementation.
// It implements the sources.Source interface, providing functionality
// for retrieving subdomains from the Certspotter API.
type Source struct{}

// Run initiates the process of retrieving subdomain information from the Certspotter API for a given domain.
//
// It constructs HTTP GET requests to the Certspotter API endpoint, processes the JSON responses,
// and streams each discovered subdomain as a sources.Result via a channel.
//
// Parameters:
//   - domain (string): The target domain for which to retrieve subdomains.
//   - cfg (*sources.Configuration): The configuration settings (which include API keys) used to authenticate with the Certspotter API.
//
// Returns:
//   - (<-chan sources.Result): A channel that asynchronously emits sources.Result values.
//     Each result is either a discovered subdomain (ResultSubdomain) or an error (ResultError)
//     encountered during the operation.
//
// The function executes the following steps:
//  1. Attempts to retrieve a random API key from the configuration's Certspotter keys.
//  2. Constructs an initial API request URL ("https://api.certspotter.com/v1/issuances") with query parameters:
//     - "domain": set to the target domain.
//     - "include_subdomains": set to "true" to include subdomains.
//     - "expand": set to "dns_names" to expand DNS name details.
//  3. Sends an HTTP GET request using the hqgohttp package and decodes the JSON response into a slice
//     of getCTLogsSearchResponse objects.
//  4. If the response is empty, terminates the operation.
//  5. Iterates over the received certificate records and, for each DNS name, checks if it is equal to the
//     target domain or a valid subdomain (ends with "." concatenated with the target domain). Valid subdomains
//     are streamed as sources.Result of type ResultSubdomain.
//  6. Retrieves the "ID" from the last record and enters a loop to paginate through additional results by
//     using the "after" parameter.
//  7. For each subsequent request, includes the "after" parameter and the Authorization header ("Bearer" token)
//     to retrieve additional certificate records. Processes each response similarly to stream subdomains.
//  8. Terminates the pagination loop when no more records are returned.
//  9. Closes the results channel upon completion.
func (source *Source) Run(domain string, cfg *sources.Configuration) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		key, err := cfg.Keys.Certspotter.PickRandom()
		if key == "" || err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		getCTLogsSearchReqURL := "https://api.certspotter.com/v1/issuances"
		getCTLogsSearchReqCFG := &hqgohttp.RequestConfiguration{
			Params: map[string]string{
				"domain":             domain,
				"include_subdomains": "true",
				"expand":             "dns_names",
			},
		}

		getCTLogsSearchRes, err := hqgohttp.Get(getCTLogsSearchReqURL, getCTLogsSearchReqCFG)
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var getCTLogsSearchResData []getCTLogsSearchResponse

		if err = json.NewDecoder(getCTLogsSearchRes.Body).Decode(&getCTLogsSearchResData); err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			getCTLogsSearchRes.Body.Close()

			return
		}

		getCTLogsSearchRes.Body.Close()

		if len(getCTLogsSearchResData) == 0 {
			return
		}

		for _, cert := range getCTLogsSearchResData {
			for _, subdomain := range cert.DNSNames {
				if subdomain != domain && !strings.HasSuffix(subdomain, "."+domain) {
					continue
				}

				result := sources.Result{
					Type:   sources.ResultSubdomain,
					Source: source.Name(),
					Value:  subdomain,
				}

				results <- result
			}
		}

		id := getCTLogsSearchResData[len(getCTLogsSearchResData)-1].ID

		for {
			getCTLogsSearchReqURL := "https://api.certspotter.com/v1/issuances"
			getCTLogsSearchReqCFG := &hqgohttp.RequestConfiguration{
				Params: map[string]string{
					"domain":             domain,
					"include_subdomains": "true",
					"expand":             "dns_names",
					"after":              id,
				},
				Headers: map[string]string{
					header.Authorization.String(): "Bearer " + key,
				},
			}

			getCTLogsSearchRes, err := hqgohttp.Get(getCTLogsSearchReqURL, getCTLogsSearchReqCFG)
			if err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: source.Name(),
					Error:  err,
				}

				results <- result

				break
			}

			var getCTLogsSearchResData []getCTLogsSearchResponse

			if err = json.NewDecoder(getCTLogsSearchRes.Body).Decode(&getCTLogsSearchResData); err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: source.Name(),
					Error:  err,
				}

				results <- result

				getCTLogsSearchRes.Body.Close()

				break
			}

			getCTLogsSearchRes.Body.Close()

			if len(getCTLogsSearchResData) == 0 {
				break
			}

			for _, cert := range getCTLogsSearchResData {
				for _, subdomain := range cert.DNSNames {
					if subdomain != domain && !strings.HasSuffix(subdomain, "."+domain) {
						continue
					}

					result := sources.Result{
						Type:   sources.ResultSubdomain,
						Source: source.Name(),
						Value:  subdomain,
					}

					results <- result
				}
			}

			id = getCTLogsSearchResData[len(getCTLogsSearchResData)-1].ID
		}
	}()

	return results
}

// Name returns the unique identifier for the Certspotter data source.
// This identifier is used for logging, debugging, and to associate results
// with the correct data source.
//
// Returns:
//   - name (string): The constant sources.CERTSPOTTER representing the Certspotter source.
func (source *Source) Name() (name string) {
	return sources.CERTSPOTTER
}
