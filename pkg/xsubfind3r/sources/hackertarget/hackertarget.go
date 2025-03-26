// Package hackertarget provides an implementation of the sources.Source interface
// for interacting with the HackerTarget API.
//
// The HackerTarget API offers host search capabilities for a given domain,
// returning subdomain information. This package defines a Source type that implements
// the Run and Name methods as specified by the sources.Source interface. The Run method sends
// a query to the HackerTarget API, processes the response using a line scanner, and streams
// discovered subdomains or errors via a channel.
package hackertarget

import (
	"bufio"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	hqgohttp "go.source.hueristiq.com/http"
)

// Source represents the HackerTarget data source implementation.
// It implements the sources.Source interface, providing functionality for retrieving
// subdomains from the HackerTarget API.
type Source struct{}

// Run initiates the process of retrieving subdomain information from the HackerTarget API for a given domain.
//
// It constructs an HTTP GET request to the HackerTarget API endpoint, processes the response line-by-line,
// applies the configured regular expression to extract subdomain matches, and streams each discovered subdomain
// as a sources.Result via a channel.
//
// Parameters:
//   - domain (string): The target domain for which to retrieve subdomains.
//   - cfg (*sources.Configuration): The configuration settings (which include the regex extractor)
//     used to parse and extract subdomains from the response.
//
// Returns:
//   - (<-chan sources.Result): A channel that asynchronously emits sources.Result values.
//     Each result is either a discovered subdomain (ResultSubdomain) or an error (ResultError)
//     encountered during the operation.
//
// The function executes the following steps:
//  1. Constructs the API request URL ("https://api.hackertarget.com/hostsearch") with the target domain as a query parameter.
//  2. Sends an HTTP GET request using the hqgohttp package.
//  3. Reads the response body using a bufio.Scanner.
//  4. For each non-empty line in the response, applies the configured regular expression to extract subdomain matches.
//  5. Streams each extracted subdomain as a sources.Result of type ResultSubdomain.
//  6. If an error occurs during scanning, streams a sources.Result of type ResultError.
//  7. Closes the results channel upon completion.
func (source *Source) Run(domain string, cfg *sources.Configuration) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		hostSearchReqURL := "https://api.hackertarget.com/hostsearch"
		hostSearchReqCFG := &hqgohttp.RequestConfiguration{
			Params: map[string]string{
				"q": domain,
			},
		}

		hostSearchRes, err := hqgohttp.Get(hostSearchReqURL, hostSearchReqCFG)
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		scanner := bufio.NewScanner(hostSearchRes.Body)

		for scanner.Scan() {
			line := scanner.Text()

			if line == "" {
				continue
			}

			match := cfg.Extractor.FindAllString(line, -1)

			for _, subdomain := range match {
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

			hostSearchRes.Body.Close()

			return
		}

		hostSearchRes.Body.Close()
	}()

	return results
}

// Name returns the unique identifier for the HackerTarget data source.
// This identifier is used for logging, debugging, and to associate results
// with the correct data source.
//
// Returns:
//   - name (string): The constant sources.HACKERTARGET representing the HackerTarget source.
func (source *Source) Name() (name string) {
	return sources.HACKERTARGET
}
