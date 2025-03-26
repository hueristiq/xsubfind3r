// Package crtsh provides an implementation of the sources.Source interface
// for interacting with the CRT.SH API.
//
// The CRT.SH API is a certificate transparency log search engine that provides
// domain name data, including discovered subdomains, by querying its public interface.
// This package defines a Source type that implements the Run and Name methods as specified
// by the sources.Source interface. The Run method sends a query to the CRT.SH API,
// processes the JSON response, and streams discovered subdomains or errors via a channel.
package crtsh

import (
	"encoding/json"
	"strings"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	hqgohttp "go.source.hueristiq.com/http"
	"go.source.hueristiq.com/http/header"
	"go.source.hueristiq.com/http/mime"
)

// getNameValuesResponse represents the structure of the JSON response returned by the CRT.SH API.
//
// It is defined as a slice of anonymous structs, where each struct contains:
//   - ID: An integer representing a unique identifier for the record.
//   - NameValue: A string that includes one or more subdomains separated by newline characters.
type getNameValuesResponse []struct {
	ID        int    `json:"id"`
	NameValue string `json:"name_value"`
}

// Source represents the CRT.SH data source implementation.
// It implements the sources.Source interface, providing functionality
// for retrieving subdomains from the CRT.SH API.
type Source struct{}

// Run initiates the process of retrieving subdomain information from the CRT.SH API for a given domain.
//
// It constructs an HTTP GET request to the CRT.SH API endpoint, decodes the JSON response,
// and streams each discovered subdomain as a sources.Result via a channel.
//
// Parameters:
//   - domain (string): The target domain for which to retrieve subdomains.
//   - _ (*sources.Configuration): The configuration settings (not used in this implementation).
//
// Returns:
//   - (<-chan sources.Result): A channel that asynchronously emits sources.Result values.
//     Each result is either a discovered subdomain (ResultSubdomain) or an error (ResultError)
//     encountered during the operation.
//
// The function executes the following steps:
//  1. Constructs the API request URL and configures the required query parameters:
//     - The query parameter "q" is set to "%." concatenated with the target domain, which searches for subdomains.
//     - The "output" parameter is set to "json" to request a JSON-formatted response.
//  2. Sets the necessary header "Content-Type" to "application/json".
//  3. Sends an HTTP GET request using the hqgohttp package.
//  4. Decodes the JSON response into a getNameValuesResponse slice.
//  5. Iterates over each record in the response:
//     - Splits the NameValue string on newline characters to extract individual subdomains.
//     - Filters the subdomains to ensure they either equal the target domain or are valid subdomains of it.
//     - Streams each valid subdomain as a sources.Result of type ResultSubdomain.
//  6. Closes the results channel upon completion.
func (source *Source) Run(domain string, _ *sources.Configuration) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		getNameValuesReqURL := "https://crt.sh"
		getNameValuesReqCFG := &hqgohttp.RequestConfiguration{
			Params: map[string]string{
				"q":      "%." + domain,
				"output": "json",
			},
			Headers: map[string]string{
				header.ContentType.String(): mime.JSON.String(),
			},
		}

		getNameValuesRes, err := hqgohttp.Get(getNameValuesReqURL, getNameValuesReqCFG)
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var getNameValuesResData getNameValuesResponse

		if err = json.NewDecoder(getNameValuesRes.Body).Decode(&getNameValuesResData); err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			getNameValuesRes.Body.Close()

			return
		}

		getNameValuesRes.Body.Close()

		for _, record := range getNameValuesResData {
			subdomains := strings.Split(record.NameValue, "\n")

			for _, subdomain := range subdomains {
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
	}()

	return results
}

// Name returns the unique identifier for the CRT.SH data source.
// This identifier is used for logging, debugging, and to associate results
// with the correct data source.
//
// Returns:
//   - name (string): The constant sources.CRTSH representing the CRT.SH source.
func (source *Source) Name() (name string) {
	return sources.CRTSH
}
