// Package censys provides an implementation of the sources.Source interface
// for interacting with the Censys API.
//
// The Censys API offers certificate transparency search capabilities for a given domain,
// returning certificate data that includes discovered subdomains. This package defines a
// Source type that implements the Run and Name methods as specified by the sources.Source
// interface. The Run method sends queries to the Censys API, processes the JSON response
// (including handling pagination via a cursor), and streams discovered subdomains or errors via a channel.
package censys

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/spf13/cast"
	hqgohttp "go.source.hueristiq.com/http"
	"go.source.hueristiq.com/http/header"
)

// certSearchResponse represents the structure of the JSON response returned by the Censys API.
//
// It contains the following fields:
//   - Code: An integer code returned by the API.
//   - Status: A string indicating the status of the API response.
//   - Error: A string containing error details if the API encountered an issue.
//   - Result: An object that contains:
//   - Query: The search query used.
//   - Total: The total number of matching records.
//   - DurationMS: The time taken by the search (in milliseconds).
//   - Hits: A slice of objects where each object represents a certificate hit.
//     Each hit includes parsed certificate details (such as validity period, subject DN, and issuer DN)
//     and a slice of Names representing discovered subdomains.
//   - Links: An object containing pagination links (Next and Prev).
type certSearchResponse struct {
	Code   int    `json:"code"`
	Status string `json:"status"`
	Error  string `json:"error"`
	Result struct {
		Query      string  `json:"query"`
		Total      float64 `json:"total"`
		DurationMS int     `json:"duration_ms"`
		Hits       []struct {
			Parsed struct {
				ValidityPeriod struct {
					NotAfter  string `json:"not_after"`
					NotBefore string `json:"not_before"`
				} `json:"validity_period"`
				SubjectDN string `json:"subject_dn"`
				IssuerDN  string `json:"issuer_dn"`
			} `json:"parsed"`
			Names             []string `json:"names"`
			FingerprintSha256 string   `json:"fingerprint_sha256"`
		} `json:"hits"`
		Links struct {
			Next string `json:"next"`
			Prev string `json:"prev"`
		} `json:"links"`
	} `json:"result"`
}

// Source represents the Censys data source implementation.
// It implements the sources.Source interface, providing functionality
// for retrieving subdomains from the Censys API.
type Source struct{}

// Run initiates the process of retrieving subdomain information from the Censys API for a given domain.
//
// It constructs an HTTP GET request to the Censys API endpoint, handles pagination via a cursor,
// decodes the JSON response, and streams each discovered subdomain as a sources.Result via a channel.
//
// Parameters:
//   - domain (string): The target domain for which to retrieve subdomains.
//   - cfg (*sources.Configuration): The configuration settings (which include API keys) used to authenticate with the Censys API.
//
// Returns:
//   - (<-chan sources.Result): A channel that asynchronously emits sources.Result values.
//     Each result is either a discovered subdomain (ResultSubdomain) or an error (ResultError)
//     encountered during the operation.
//
// The function executes the following steps:
//  1. Attempts to retrieve a random API key from the configuration's Censys keys.
//  2. Initializes a pagination cursor variable and a page counter.
//  3. Enters a loop to send HTTP GET requests to the Censys API, including the pagination cursor if present:
//     a. Constructs the API request URL ("https://search.censys.io/api/v2/certificates/search")
//     and sets query parameters including the search query (q) and the number of results per page.
//     b. Sets the "Authorization" header using Basic authentication with the base64-encoded API key.
//     c. Sends the HTTP GET request using the hqgohttp package.
//     d. Decodes the JSON response into a certSearchResponse struct.
//     e. If the API response contains an error message, streams an error result and terminates the loop.
//     f. Iterates over each certificate hit and, for each name in the hit, streams it as a sources.Result
//     of type ResultSubdomain.
//     g. Updates the pagination cursor from the Result.Links.Next field and increments the page counter.
//     h. Terminates the loop if no further pagination cursor is provided or if the maximum number of pages is reached.
//  4. Closes the results channel upon completion.
func (source *Source) Run(domain string, cfg *sources.Configuration) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		key, err := cfg.Keys.Censys.PickRandom()
		if key == "" || err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		page := 1
		cursor := ""

		certSearchReqURL := "https://search.censys.io/api/v2/certificates/search"

		for {
			certSearchReqCFG := &hqgohttp.RequestConfiguration{
				Params: map[string]string{
					"q":        domain,
					"per_page": cast.ToString(maxPerPage),
				},
				Headers: map[string]string{
					header.Authorization.String(): "Basic " + base64.StdEncoding.EncodeToString([]byte(key)),
				},
			}

			if cursor != "" {
				certSearchReqCFG.Params["cursor"] = cursor
			}

			certSearchRes, err := hqgohttp.Get(certSearchReqURL, certSearchReqCFG)
			if err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: source.Name(),
					Error:  err,
				}

				results <- result

				return
			}

			var certSearchResData certSearchResponse

			if err = json.NewDecoder(certSearchRes.Body).Decode(&certSearchResData); err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: source.Name(),
					Error:  err,
				}

				results <- result

				certSearchRes.Body.Close()

				return
			}

			certSearchRes.Body.Close()

			if certSearchResData.Error != "" {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: source.Name(),
					Error:  fmt.Errorf("%w: %s, %s", errStatic, certSearchResData.Status, certSearchResData.Error),
				}

				results <- result

				return
			}

			for _, hit := range certSearchResData.Result.Hits {
				for _, name := range hit.Names {
					result := sources.Result{
						Type:   sources.ResultSubdomain,
						Source: source.Name(),
						Value:  name,
					}

					results <- result
				}
			}

			cursor = certSearchResData.Result.Links.Next

			if cursor == "" || page >= maxCensysPages {
				break
			}

			page++
		}
	}()

	return results
}

// Name returns the unique identifier for the Censys data source.
// This identifier is used for logging, debugging, and to associate results
// with the correct data source.
//
// Returns:
//   - name (string): The constant sources.CENSYS representing the Censys source.
func (source *Source) Name() (name string) {
	return sources.CENSYS
}

const (
	maxCensysPages = 10
	maxPerPage     = 100
)

// errStatic is a sentinel error used to prepend error messages when the Censys API response
// contains error details.
var errStatic = errors.New("something went wrong")
