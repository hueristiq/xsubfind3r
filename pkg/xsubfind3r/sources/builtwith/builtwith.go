// Package builtwith provides an implementation of the sources.Source interface
// for interacting with the BuiltWith API.
//
// The BuiltWith API offers detailed information on the technologies used by a domain,
// including subdomain discovery. This package defines a Source type that implements the
// Run and Name methods as specified by the sources.Source interface. The Run method sends
// a query to the BuiltWith API, processes the JSON response, and streams discovered subdomains
// or errors via a channel.
package builtwith

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	hqgohttp "go.source.hueristiq.com/http"
)

// getDomainInfoResponse represents the structure of the JSON response returned by the BuiltWith API.
//
// It contains two primary components:
//   - Results: A slice of result objects, each of which includes technology paths with domain,
//     URL, and subdomain information.
//   - Errors:  A slice of error objects, each containing details about an error encountered during the API request.
type getDomainInfoResponse struct {
	Results []struct {
		Result struct {
			Paths []struct {
				Domain    string `json:"Domain"`
				URL       string `json:"Url"`
				SubDomain string `json:"SubDomain"`
			} `json:"Paths"`
		} `json:"Result"`
	} `json:"Results"`
	Errors []struct {
		Lookup  string `json:"Lookup"`
		Message string `json:"Message"`
	} `json:"Errors"`
}

// Source represents the BuiltWith data source implementation.
// It implements the sources.Source interface, providing functionality
// for retrieving subdomains from the BuiltWith API.
type Source struct{}

// Run initiates the process of retrieving subdomain information from the BuiltWith API for a given domain.
//
// It constructs an HTTP GET request to the BuiltWith API endpoint, decodes the JSON response,
// and streams each discovered subdomain as a sources.Result via a channel.
//
// Parameters:
//   - domain (string): The target domain for which to retrieve subdomains.
//   - cfg (*sources.Configuration): The configuration settings (which include API keys) used to authenticate with the BuiltWith API.
//
// Returns:
//   - (<-chan sources.Result): A channel that asynchronously emits sources.Result values.
//     Each result is either a discovered subdomain (ResultSubdomain) or an error (ResultError)
//     encountered during the operation.
//
// The function executes the following steps:
//  1. Attempts to retrieve a random API key from the configuration's BuiltWith keys.
//  2. Constructs the API request URL and configures the required parameters, including the API key and query options.
//  3. Sends the HTTP GET request using the hqgohttp package.
//  4. Decodes the JSON response into a getDomainInfoResponse struct.
//  5. Checks for errors in the response; if any are found, streams each error as a sources.Result of type ResultError.
//  6. Iterates over the results and their associated paths, concatenating SubDomain and Domain to form full subdomain strings,
//     and streams each discovered subdomain as a sources.Result of type ResultSubdomain.
//  7. Closes the results channel upon completion.
func (source *Source) Run(domain string, cfg *sources.Configuration) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		key, err := cfg.Keys.BuiltWith.PickRandom()
		if key == "" || err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		getDomainInfoReqURL := "https://api.builtwith.com/v21/api.json"
		getDomainInfoReqCFG := &hqgohttp.RequestConfiguration{
			Params: map[string]string{
				"KEY":      key,
				"HIDETEXT": "yes",
				"HIDEDL":   "yes",
				"NOLIVE":   "yes",
				"NOMETA":   "yes",
				"NOPII":    "yes",
				"NOATTR":   "yes",
				"LOOKUP":   domain,
			},
		}

		getDomainInfoRes, err := hqgohttp.Get(getDomainInfoReqURL, getDomainInfoReqCFG)
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var getDomainInfoResData getDomainInfoResponse

		if err = json.NewDecoder(getDomainInfoRes.Body).Decode(&getDomainInfoResData); err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			getDomainInfoRes.Body.Close()

			return
		}

		getDomainInfoRes.Body.Close()

		if len(getDomainInfoResData.Errors) > 0 {
			for _, entry := range getDomainInfoResData.Errors {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: source.Name(),
					Error:  fmt.Errorf("%w: %s", errStatic, entry.Message),
				}

				results <- result
			}

			return
		}

		for _, item := range getDomainInfoResData.Results {
			for _, path := range item.Result.Paths {
				result := sources.Result{
					Type:   sources.ResultSubdomain,
					Source: source.Name(),
					Value:  path.SubDomain + "." + path.Domain,
				}

				results <- result
			}
		}
	}()

	return results
}

// Name returns the unique identifier for the BuiltWith data source.
// This identifier is used for logging, debugging, and to associate results
// with the correct data source.
//
// Returns:
//   - name (string): The constant sources.BUILTWITH representing the BuiltWith source.
func (source *Source) Name() (name string) {
	return sources.BUILTWITH
}

// errStatic is a sentinel error used to prepend error messages when the BuiltWith API response
// contains error messages.
var errStatic = errors.New("something went wrong")
