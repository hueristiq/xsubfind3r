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

// Name returns the unique identifier for the data source.
// This identifier is used for logging, debugging, and associating results with the correct data source.
//
// Returns:
//   - name (string): The unique identifier for the data source.
func (source *Source) Name() (name string) {
	return sources.BUILTWITH
}

// errStatic is a sentinel error used to prepend error messages when the BuiltWith API response
// contains error messages.
var errStatic = errors.New("something went wrong")
