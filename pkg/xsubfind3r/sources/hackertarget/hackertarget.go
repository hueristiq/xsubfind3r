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
// It implements the sources.Source interface, providing functionality
// for retrieving subdomains from the HackerTarget API.
type Source struct{}

// Run initiates the process of retrieving subdomain information from the HackerTarget API for a given domain.
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

// Name returns the unique identifier for the data source.
// This identifier is used for logging, debugging, and associating results with the correct data source.
//
// Returns:
//   - name (string): The unique identifier for the data source.
func (source *Source) Name() (name string) {
	return sources.HACKERTARGET
}
