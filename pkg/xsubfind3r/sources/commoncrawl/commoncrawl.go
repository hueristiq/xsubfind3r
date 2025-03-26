// Package commoncrawl provides an implementation of the sources.Source interface
// for interacting with the Common Crawl index.
//
// The Common Crawl index provides archived web data that can be used to discover
// subdomains for a given domain by searching historical URLs. This package defines a
// Source type that implements the Run and Name methods as specified by the sources.Source
// interface. The Run method retrieves index metadata, filters for recent indexes, queries
// the index for URLs matching the target domain, extracts subdomains using a provided regular
// expression, and streams discovered subdomains or errors via a channel.
package commoncrawl

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/spf13/cast"
	hqgohttp "go.source.hueristiq.com/http"
	"go.source.hueristiq.com/http/header"
)

// getIndexesResponse represents the structure of the JSON response returned by
// the Common Crawl index metadata endpoint.
//
// It is defined as a slice of anonymous structs, where each struct contains:
//   - ID: A string identifier for the index.
//   - Name: The name of the index.
//   - TimeGate: A URL for time-based redirection.
//   - CDXAPI: A string containing the API endpoint URL for that index.
//   - From: A string representing the start date of the index.
//   - To: A string representing the end date of the index.
type getIndexesResponse []struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	TimeGate string `json:"timegate"`
	CDXAPI   string `json:"cdx-api"`
	From     string `json:"from"`
	To       string `json:"to"`
}

// getPaginationResponse represents the structure of the JSON response that provides
// pagination information for a Common Crawl index query.
//
// It contains the following fields:
//   - Blocks: The number of data blocks available.
//   - PageSize: The number of records per page.
//   - Pages: The total number of pages available for the query.
type getPaginationResponse struct {
	Blocks   uint `json:"blocks"`
	PageSize uint `json:"pageSize"`
	Pages    uint `json:"pages"`
}

// getURLsResponse represents the structure of each JSON record returned when querying
// a Common Crawl index for URLs.
//
// It contains the following fields:
//   - URL: A string representing a discovered URL.
//   - Error: A string describing an error encountered for the record, if any.
type getURLsResponse struct {
	URL   string `json:"url"`
	Error string `json:"error"`
}

// Source represents the Common Crawl data source implementation.
// It implements the sources.Source interface, providing functionality
// for retrieving subdomains from the Common Crawl index.
type Source struct{}

// Run initiates the process of retrieving subdomain information from the Common Crawl index for a given domain.
//
// Parameters:
//   - domain (string): The target domain for which to retrieve subdomains.
//   - cfg (*sources.Configuration): The configuration settings (which include API keys and the regular expression extractor)
//     used to authenticate with the Common Crawl API and to extract subdomains from URLs.
//
// Returns:
//   - (<-chan sources.Result): A channel that asynchronously emits sources.Result values.
//     Each result is either a discovered subdomain (ResultSubdomain) or an error (ResultError)
//     encountered during the operation.
//
// The function executes the following steps:
//  1. Sends an HTTP GET request to retrieve index metadata from the Common Crawl collection info endpoint.
//  2. Filters the retrieved indexes to select those corresponding to the current year and the past few years (up to 5 years back).
//  3. For each selected index API URL, constructs a query with parameters to search for URLs matching the target domain.
//  4. Sends an HTTP GET request to retrieve pagination information (number of pages) for the query.
//  5. For each page, sends an HTTP GET request to retrieve URLs.
//  6. Uses a buffered scanner to iterate over each line of the response, decoding each JSON record.
//  7. Checks for errors in each record and, if none, extracts subdomains from the URL using the provided regular expression (cfg.Extractor).
//  8. Streams each discovered subdomain as a sources.Result of type ResultSubdomain or any encountered error as a ResultError.
//  9. Closes the results channel upon completion.
func (source *Source) Run(domain string, cfg *sources.Configuration) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		getIndexesReqURL := "https://index.commoncrawl.org/collinfo.json"

		getIndexesRes, err := hqgohttp.Get(getIndexesReqURL)
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var getIndexesResData getIndexesResponse

		if err = json.NewDecoder(getIndexesRes.Body).Decode(&getIndexesResData); err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			getIndexesRes.Body.Close()

			return
		}

		getIndexesRes.Body.Close()

		year := time.Now().Year()
		years := make([]string, 0)
		maxYearsBack := 5

		for i := range maxYearsBack {
			years = append(years, strconv.Itoa(year-i))
		}

		searchIndexes := make(map[string]string)

		for _, year := range years {
			for _, CCIndex := range getIndexesResData {
				if strings.Contains(CCIndex.ID, year) {
					if _, ok := searchIndexes[year]; !ok {
						searchIndexes[year] = CCIndex.CDXAPI

						break
					}
				}
			}
		}

		for _, CCIndexAPI := range searchIndexes {
			getPaginationReqCFG := &hqgohttp.RequestConfiguration{
				Headers: map[string]string{
					header.Host.String(): "index.commoncrawl.org",
				},
				Params: map[string]string{
					"url":          "*." + domain + "/*",
					"output":       "json",
					"fl":           "url",
					"showNumPages": "true",
				},
			}

			getPaginationRes, err := hqgohttp.Get(CCIndexAPI, getPaginationReqCFG)
			if err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: source.Name(),
					Error:  err,
				}

				results <- result

				continue
			}

			var getPaginationResData getPaginationResponse

			if err = json.NewDecoder(getPaginationRes.Body).Decode(&getPaginationResData); err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: source.Name(),
					Error:  err,
				}

				results <- result

				getPaginationRes.Body.Close()

				continue
			}

			getPaginationRes.Body.Close()

			if getPaginationResData.Pages < 1 {
				continue
			}

			for page := range getPaginationResData.Pages {
				getURLsReqCFG := &hqgohttp.RequestConfiguration{
					Headers: map[string]string{
						header.Host.String(): "index.commoncrawl.org",
					},
					Params: map[string]string{
						"url":    "*." + domain + "/*",
						"output": "json",
						"fl":     "url",
						"page":   cast.ToString(page),
					},
				}

				getURLsRes, err := hqgohttp.Get(CCIndexAPI, getURLsReqCFG)
				if err != nil {
					result := sources.Result{
						Type:   sources.ResultError,
						Source: source.Name(),
						Error:  err,
					}

					results <- result

					continue
				}

				scanner := bufio.NewScanner(getURLsRes.Body)

				for scanner.Scan() {
					var getURLsResData getURLsResponse

					if err = json.Unmarshal(scanner.Bytes(), &getURLsResData); err != nil {
						result := sources.Result{
							Type:   sources.ResultError,
							Source: source.Name(),
							Error:  err,
						}

						results <- result

						continue
					}

					if getURLsResData.Error != "" {
						result := sources.Result{
							Type:   sources.ResultError,
							Source: source.Name(),
							Error:  fmt.Errorf("%w: %s", errStatic, getURLsResData.Error),
						}

						results <- result

						continue
					}

					subdomains := cfg.Extractor.FindAllString(getURLsResData.URL, -1)

					for _, subdomain := range subdomains {
						result := sources.Result{
							Type:   sources.ResultSubdomain,
							Source: source.Name(),
							Value:  subdomain,
						}

						results <- result
					}
				}

				if err = scanner.Err(); err != nil {
					result := sources.Result{
						Type:   sources.ResultError,
						Source: source.Name(),
						Error:  err,
					}

					results <- result

					getURLsRes.Body.Close()

					continue
				}

				getURLsRes.Body.Close()
			}
		}
	}()

	return results
}

// Name returns the unique identifier for the Common Crawl data source.
// This identifier is used for logging, debugging, and to associate results
// with the correct data source.
//
// Returns:
//   - name (string): The constant sources.COMMONCRAWL representing the Common Crawl source.
func (source *Source) Name() (name string) {
	return sources.COMMONCRAWL
}

// errStatic is a sentinel error used to prepend error messages when a
// record-specific error is encountered in the Common Crawl responses.
var errStatic = errors.New("something went wrong")
