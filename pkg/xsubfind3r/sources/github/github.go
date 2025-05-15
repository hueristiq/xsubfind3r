// Package github provides an implementation of the sources.Source interface
// for interacting with the GitHub API.
//
// The GitHub API can be used to search for code related to a given domain, where
// subdomain information may be present in the code or in text matches.
// This package defines a Source type that implements the Run, Enumerate, and Name methods
// as specified by the sources.Source interface. The Run method initiates a code search query,
// and the Enumerate method handles processing of the search results, including pagination,
// rate limiting, and extraction of subdomains from both raw file content and text matches.
package github

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	hqgohttp "github.com/hueristiq/hq-go-http"
	"github.com/hueristiq/hq-go-http/header"
	"github.com/hueristiq/hq-go-http/status"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/spf13/cast"
)

// codeSearchResponse represents the structure of the JSON response returned by the GitHub code search API.
//
// It contains the total count of matching records and a slice of items where each item
// represents a code search result. Each item includes the repository file name, the HTML URL for the file,
// and any text matches found in the file.
type codeSearchResponse struct {
	TotalCount int `json:"total_count"`
	Items      []struct {
		Name        string `json:"name"`
		HTMLURL     string `json:"html_url"`
		TextMatches []struct {
			Fragment string `json:"fragment"`
		} `json:"text_matches"`
	} `json:"items"`
}

// Source represents the GitHub data source implementation.
// It implements the sources.Source interface, providing functionality
// for retrieving subdomains by querying GitHub code search results.
type Source struct{}

// Run initiates the process of retrieving subdomain information from GitHub for a given domain.
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

		if len(cfg.Keys.GitHub) == 0 {
			return
		}

		tokens := NewTokenManager(cfg.Keys.GitHub)

		searchReqURL := fmt.Sprintf(
			"https://api.github.com/search/code?per_page=100&q=%q&sort=created&order=asc",
			domain,
		)

		source.Enumerate(searchReqURL, tokens, cfg, results)
	}()

	return results
}

// Enumerate processes GitHub code search results by sending HTTP GET requests to the provided search URL,
// handling pagination via the Link header, and extracting subdomains from raw file content and text matches.
//
// Parameters:
//   - searchReqURL (string): The URL for the GitHub code search API request.
//   - tokens (*Tokens): A token manager containing GitHub API tokens to handle rate limiting.
//   - cfg (*sources.Configuration): The configuration settings used for authentication and regex extraction.
//   - results (chan sources.Result): A channel to stream discovered subdomains or errors.
func (source *Source) Enumerate(searchReqURL string, tokens *Tokens, cfg *sources.Configuration, results chan sources.Result) {
	token := tokens.Get()

	if token.RetryAfter > 0 {
		if len(tokens.pool) == 1 {
			time.Sleep(time.Duration(token.RetryAfter) * time.Second)
		} else {
			token = tokens.Get()
		}
	}

	codeSearchResCFG := &hqgohttp.RequestConfiguration{
		Headers: []hqgohttp.Header{
			hqgohttp.NewHeader(header.Accept.String(), "application/vnd.github.v3.text-match+json", hqgohttp.HeaderModeSet),
			hqgohttp.NewHeader(header.Authorization.String(), "token "+token.Hash, hqgohttp.HeaderModeSet),
		},
	}

	codeSearchRes, err := hqgohttp.Get(searchReqURL, codeSearchResCFG)

	isForbidden := codeSearchRes != nil && codeSearchRes.StatusCode == status.Forbidden.Int()

	if err != nil && !isForbidden {
		result := sources.Result{
			Type:   sources.ResultError,
			Source: source.Name(),
			Error:  err,
		}

		results <- result

		return
	}

	ratelimitRemaining := cast.ToInt64(
		codeSearchRes.Header.Get(header.XRatelimitRemaining.String()),
	)
	if isForbidden && ratelimitRemaining == 0 {
		retryAfterSeconds := cast.ToInt64(codeSearchRes.Header.Get(header.RetryAfter.String()))

		tokens.setCurrentTokenExceeded(retryAfterSeconds)

		source.Enumerate(searchReqURL, tokens, cfg, results)
	}

	var codeSearchResData codeSearchResponse

	if err = json.NewDecoder(codeSearchRes.Body).Decode(&codeSearchResData); err != nil {
		result := sources.Result{
			Type:   sources.ResultError,
			Source: source.Name(),
			Error:  err,
		}

		results <- result

		codeSearchRes.Body.Close()

		return
	}

	codeSearchRes.Body.Close()

	for _, item := range codeSearchResData.Items {
		getRawContentReqURL := strings.ReplaceAll(
			item.HTMLURL,
			"https://github.com/",
			"https://raw.githubusercontent.com/",
		)
		getRawContentReqURL = strings.ReplaceAll(getRawContentReqURL, "/blob/", "/")

		var getRawContentRes *http.Response

		getRawContentRes, err = hqgohttp.Get(getRawContentReqURL)
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			continue
		}

		if getRawContentRes.StatusCode != status.OK.Int() {
			continue
		}

		scanner := bufio.NewScanner(getRawContentRes.Body)

		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			subdomains := cfg.Extractor.FindAllString(line, -1)

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

			getRawContentRes.Body.Close()

			return
		}

		getRawContentRes.Body.Close()

		for _, match := range item.TextMatches {
			subdomains := cfg.Extractor.FindAllString(match.Fragment, -1)

			for _, subdomain := range subdomains {
				result := sources.Result{
					Type:   sources.ResultSubdomain,
					Source: source.Name(),
					Value:  subdomain,
				}

				results <- result
			}
		}
	}

	links := header.ParseLinkHeader(codeSearchRes.Header.Get(header.Link.String()))

	for _, link := range links {
		if link.Rel == "next" {
			nextURL, err := url.QueryUnescape(link.URL)
			if err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: source.Name(),
					Error:  err,
				}

				results <- result

				return
			}

			source.Enumerate(nextURL, tokens, cfg, results)
		}
	}
}

// Name returns the unique identifier for the data source.
// This identifier is used for logging, debugging, and associating results with the correct data source.
//
// Returns:
//   - name (string): The unique identifier for the data source.
func (source *Source) Name() string {
	return sources.GITHUB
}
