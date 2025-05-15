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

	hqgohttp "github.com/hueristiq/hq-go-http"
	"github.com/hueristiq/hq-go-http/header"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
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
				Headers: []hqgohttp.Header{
					hqgohttp.NewHeader(header.Authorization.String(), "Bearer "+key, hqgohttp.HeaderModeSet),
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

// Name returns the unique identifier for the data source.
// This identifier is used for logging, debugging, and associating results with the correct data source.
//
// Returns:
//   - name (string): The unique identifier for the data source.
func (source *Source) Name() (name string) {
	return sources.CERTSPOTTER
}
