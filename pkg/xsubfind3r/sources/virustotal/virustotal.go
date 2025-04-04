// Package virustotal provides an implementation of the sources.Source interface
// for interacting with the VirusTotal API.
//
// The VirusTotal API offers subdomain discovery for a given domain by returning
// subdomain data and pagination details. This package defines a Source type that implements
// the Run and Name methods as specified by the sources.Source interface. The Run method sends
// queries to the VirusTotal API, processes the JSON response, and streams discovered subdomains
// or errors via a channel.
package virustotal

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	hqgohttp "go.source.hueristiq.com/http"
)

// getSubdomainsResponse represents the structure of the JSON response returned by the VirusTotal API.
//
// It contains the following fields:
//   - Error: An object containing error details if the API encountered an error.
//   - Data: A slice of objects where each object represents a discovered subdomain.
//     Each object contains an ID (the subdomain), a Type, and associated Links.
//   - Meta: A metadata object containing a Cursor field used for pagination.
type getSubdomainsResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	Data []struct {
		ID    string `json:"id"`
		Type  string `json:"type"`
		Links struct {
			Self string `json:"self"`
		} `json:"links"`
	} `json:"data"`
	Meta struct {
		Cursor string `json:"cursor"`
	} `json:"meta"`
}

// Source represents the VirusTotal data source implementation.
// It implements the sources.Source interface, providing functionality
// for retrieving subdomains from the VirusTotal API.
type Source struct{}

// Run initiates the process of retrieving subdomain information from the VirusTotal API for a given domain.
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

		key, err := cfg.Keys.VirusTotal.PickRandom()
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var cursor string

		for {
			getSubdomainsReqURL := fmt.Sprintf(
				"https://www.virustotal.com/api/v3/domains/%s/subdomains",
				domain,
			)
			getSubdomainsReqCFG := &hqgohttp.RequestConfiguration{
				Params: map[string]string{
					"limit": "40",
				},
				Headers: map[string]string{
					"x-apikey": key,
				},
			}

			if cursor != "" {
				getSubdomainsReqCFG.Params["cursor"] = cursor
			}

			getSubdomainsRes, err := hqgohttp.Get(getSubdomainsReqURL, getSubdomainsReqCFG)
			if err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: source.Name(),
					Error:  err,
				}

				results <- result

				break
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

				break
			}

			getSubdomainsRes.Body.Close()

			if getSubdomainsResData.Error.Message != "" {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: source.Name(),
					Error: fmt.Errorf(
						"%w: %s, %s",
						errStatic,
						getSubdomainsResData.Error.Code,
						getSubdomainsResData.Error.Message,
					),
				}

				results <- result

				break
			}

			for _, record := range getSubdomainsResData.Data {
				subdomain := record.ID

				result := sources.Result{
					Type:   sources.ResultSubdomain,
					Source: source.Name(),
					Value:  subdomain,
				}

				results <- result
			}

			cursor = getSubdomainsResData.Meta.Cursor

			if cursor == "" {
				break
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
	return sources.VIRUSTOTAL
}

// errStatic is a sentinel error used to prepend error messages when the VirusTotal API response
// contains error details.
var errStatic = errors.New("something went wrong")
