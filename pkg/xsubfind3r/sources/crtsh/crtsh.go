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

	hqgohttp "github.com/hueristiq/hq-go-http"
	hqgohttpheader "github.com/hueristiq/hq-go-http/header"
	hqgohttpmime "github.com/hueristiq/hq-go-http/mime"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
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
// Parameters:
//   - domain (string): The target domain for which to retrieve subdomains.
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

		getNameValuesReqURL := "https://crt.sh"
		getNameValuesReqCFG := &hqgohttp.RequestConfiguration{
			Params: map[string]string{
				"q":      "%." + domain,
				"output": "json",
			},
			Headers: []hqgohttp.Header{
				hqgohttp.NewSetHeader(hqgohttpheader.ContentType.String(), hqgohttpmime.JSON.String()),
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

// Name returns the unique identifier for the data source.
// This identifier is used for logging, debugging, and associating results with the correct data source.
//
// Returns:
//   - name (string): The unique identifier for the data source.
func (source *Source) Name() (name string) {
	return sources.CRTSH
}
