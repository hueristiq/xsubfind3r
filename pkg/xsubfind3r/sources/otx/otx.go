// Package otx provides an implementation of the sources.Source interface
// for interacting with the OTX (Open Threat Exchange) API.
//
// The OTX API offers passive DNS data for a given domain, which includes historical
// DNS records and related information. This package defines a Source type that implements
// the Run and Name methods as specified by the sources.Source interface. The Run method sends
// a query to the OTX API, processes the JSON response, and streams discovered subdomains or errors
// via a channel.
package otx

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	hqgohttp "go.source.hueristiq.com/http"
)

// getPassiveDNSResponse represents the structure of the JSON response returned by the OTX API.
//
// It contains the following fields:
//   - Detail: Additional details provided in the API response.
//   - Error:  A string containing error information if the request encountered an issue.
//   - PassiveDNS: A slice of passive DNS records, each containing a hostname representing a discovered subdomain.
type getPassiveDNSResponse struct {
	Detail     string `json:"detail"`
	Error      string `json:"error"`
	PassiveDNS []struct {
		Hostname string `json:"hostname"`
	} `json:"passive_dns"`
}

// Source represents the OTX data source implementation.
// It implements the sources.Source interface, providing functionality
// for retrieving passive DNS data (subdomains) from the OTX API.
type Source struct{}

// Run initiates the process of retrieving passive DNS information from the OTX API for a given domain.
//
// It constructs an HTTP GET request to the OTX API endpoint, decodes the JSON response,
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
//  1. Constructs the API request URL using the target domain.
//  2. Sends an HTTP GET request using the hqgohttp package.
//  3. Decodes the JSON response into a getPassiveDNSResponse struct.
//  4. Checks if the response contains an error; if so, streams an error result.
//  5. Iterates over the passive DNS records, filtering results to ensure they belong to the target domain,
//     and streams each valid subdomain as a sources.Result of type ResultSubdomain.
//  6. Closes the results channel upon completion.
func (source *Source) Run(domain string, _ *sources.Configuration) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		getPassiveDNSReqURL := fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/domain/%s/passive_dns", domain)

		getPassiveDNSRes, err := hqgohttp.Get(getPassiveDNSReqURL)
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var getPassiveDNSResData getPassiveDNSResponse

		if err = json.NewDecoder(getPassiveDNSRes.Body).Decode(&getPassiveDNSResData); err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			getPassiveDNSRes.Body.Close()

			return
		}

		getPassiveDNSRes.Body.Close()

		if getPassiveDNSResData.Error != "" {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  fmt.Errorf("%w: %s, %s", errStatic, getPassiveDNSResData.Detail, getPassiveDNSResData.Error),
			}

			results <- result

			return
		}

		for _, record := range getPassiveDNSResData.PassiveDNS {
			subdomain := record.Hostname

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
	}()

	return results
}

// Name returns the unique identifier for the OTX data source.
// This identifier is used for logging, debugging, and to associate results
// with the correct data source.
//
// Returns:
//   - name (string): The constant sources.OPENTHREATEXCHANGE representing the OTX source.
func (source *Source) Name() (name string) {
	return sources.OPENTHREATEXCHANGE
}

// errStatic is a sentinel error used to prepend error messages when the OTX API response
// contains error details.
var errStatic = errors.New("something went wrong")
