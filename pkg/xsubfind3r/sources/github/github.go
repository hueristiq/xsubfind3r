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
	hqgohttpheader "github.com/hueristiq/hq-go-http/header"
	hqgohttpheaderutils "github.com/hueristiq/hq-go-http/header/utils"
	hqgohttpstatus "github.com/hueristiq/hq-go-http/status"
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
//
// Fields:
//   - tokens (*Tokens): A token manager containing GitHub API tokens to handle rate limiting.
type Source struct {
	tokens *Tokens
}

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
func (s *Source) Run(domain string, cfg *sources.Configuration) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		if len(cfg.Keys.GitHub) == 0 {
			return
		}

		s.tokens = NewTokenManager(cfg.Keys.GitHub)

		searchReqURL := fmt.Sprintf(
			"https://api.github.com/search/code?per_page=100&q=%q&sort=created&order=asc",
			domain,
		)

		s.Enumerate(searchReqURL, cfg, results)
	}()

	return results
}

// Enumerate processes GitHub code search results by sending HTTP GET requests to the provided search URL,
// handling pagination via the Link header, and extracting subdomains from raw file content and text matches.
//
// Parameters:
//   - searchReqURL (string): The URL for the GitHub code search API request.
//   - cfg (*sources.Configuration): The configuration settings used for authentication and regex extraction.
//   - results (chan sources.Result): A channel to stream discovered subdomains or errors.
func (s *Source) Enumerate(searchReqURL string, cfg *sources.Configuration, results chan sources.Result) {
	token := s.tokens.Get()

	if token.RetryAfter > 0 {
		if len(s.tokens.pool) == 1 {
			time.Sleep(time.Duration(token.RetryAfter) * time.Second)
		} else {
			token = s.tokens.Get()
		}
	}

	codeSearchResCFG := &hqgohttp.RequestConfiguration{
		Headers: []hqgohttp.Header{
			hqgohttp.NewSetHeader(hqgohttpheader.Accept.String(), "application/vnd.github.v3.text-match+json"),
			hqgohttp.NewSetHeader(hqgohttpheader.Authorization.String(), "token "+token.Hash),
		},
	}

	codeSearchRes, err := hqgohttp.Get(searchReqURL, codeSearchResCFG)

	isForbidden := codeSearchRes != nil && codeSearchRes.StatusCode == hqgohttpstatus.Forbidden.Int()

	if err != nil && !isForbidden {
		result := sources.Result{
			Type:   sources.ResultError,
			Source: s.Name(),
			Error:  err,
		}

		results <- result

		return
	}

	ratelimitRemaining := cast.ToInt64(
		codeSearchRes.Header.Get(hqgohttpheader.XRatelimitRemaining.String()),
	)

	if isForbidden && ratelimitRemaining == 0 {
		retryAfterSeconds := cast.ToInt64(codeSearchRes.Header.Get(hqgohttpheader.RetryAfter.String()))

		s.tokens.setCurrentTokenExceeded(retryAfterSeconds)

		s.Enumerate(searchReqURL, cfg, results)
	}

	var codeSearchResData codeSearchResponse

	if err = json.NewDecoder(codeSearchRes.Body).Decode(&codeSearchResData); err != nil {
		result := sources.Result{
			Type:   sources.ResultError,
			Source: s.Name(),
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
				Source: s.Name(),
				Error:  err,
			}

			results <- result

			continue
		}

		if getRawContentRes.StatusCode != hqgohttpstatus.OK.Int() {
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
					Source: s.Name(),
					Value:  subdomain,
				}

				results <- result
			}
		}

		if err = scanner.Err(); err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
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
					Source: s.Name(),
					Value:  subdomain,
				}

				results <- result
			}
		}
	}

	links := hqgohttpheaderutils.ParseLinkHeaderValue(codeSearchRes.Header.Get(hqgohttpheader.Link.String()))

	for _, link := range links {
		if link.Rel == "next" {
			nextURL, err := url.QueryUnescape(link.URL)
			if err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: s.Name(),
					Error:  err,
				}

				results <- result

				return
			}

			s.Enumerate(nextURL, cfg, results)
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

type Token struct {
	Hash         string
	RetryAfter   int64
	ExceededTime time.Time
}

type Tokens struct {
	current int
	pool    []Token
}

func NewTokenManager(keys []string) *Tokens {
	pool := []Token{}

	for _, key := range keys {
		t := Token{Hash: key, ExceededTime: time.Time{}, RetryAfter: 0}

		pool = append(pool, t)
	}

	return &Tokens{
		current: 0,
		pool:    pool,
	}
}

func (r *Tokens) setCurrentTokenExceeded(retryAfter int64) {
	if r.current >= len(r.pool) {
		r.current %= len(r.pool)
	}

	if r.pool[r.current].RetryAfter == 0 {
		r.pool[r.current].ExceededTime = time.Now()
		r.pool[r.current].RetryAfter = retryAfter
	}
}

func (r *Tokens) Get() *Token {
	resetExceededTokens(r)

	if r.current >= len(r.pool) {
		r.current %= len(r.pool)
	}

	result := &r.pool[r.current]

	r.current++

	return result
}

func resetExceededTokens(r *Tokens) {
	for i, token := range r.pool {
		if token.RetryAfter > 0 {
			if int64(time.Since(token.ExceededTime)/time.Second) > token.RetryAfter {
				r.pool[i].ExceededTime = time.Time{}
				r.pool[i].RetryAfter = 0
			}
		}
	}
}
