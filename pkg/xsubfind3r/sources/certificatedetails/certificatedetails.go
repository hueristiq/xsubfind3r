// Package certificatedetails provides an implementation of the sources.Source interface
// for interacting with the CertificateDetails website.
//
// The CertificateDetails service provides certificate transparency data for a given domain,
// which can be used to discover subdomains. This package defines a Source type that implements
// the Run and Name methods as specified by the sources.Source interface. The Run method sends
// a query to the CertificateDetails website, processes the response line-by-line, extracts
// subdomains using a regular expression provided in the configuration, and streams discovered
// subdomains or errors via a channel.
package certificatedetails

import (
	"bufio"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	hqgohttp "go.source.hueristiq.com/http"
	"go.source.hueristiq.com/http/status"
)

// Source represents the CertificateDetails data source implementation.
// It implements the sources.Source interface, providing functionality
// for retrieving subdomains from the CertificateDetails website.
type Source struct{}

// Run initiates the process of retrieving subdomain information from the CertificateDetails website for a given domain.
//
// It constructs an HTTP GET request to the CertificateDetails URL for the target domain,
// processes the response using a buffered scanner, applies the configured regular expression
// to extract subdomain matches, and streams each discovered subdomain as a sources.Result via a channel.
//
// Parameters:
//   - domain (string): The target domain for which to retrieve subdomains.
//   - cfg (*sources.Configuration): The configuration settings (which include the regular expression extractor)
//     used to parse and extract subdomains from the response.
//
// Returns:
//   - (<-chan sources.Result): A channel that asynchronously emits sources.Result values.
//     Each result is either a discovered subdomain (ResultSubdomain) or an error (ResultError)
//     encountered during the operation.
//
// The function executes the following steps:
//  1. Constructs the CertificateDetails request URL using the target domain.
//  2. Sends an HTTP GET request using the hqgohttp package.
//  3. If an error occurs during the request (and the status code is not "Not Found"),
//     streams an error result and terminates the operation.
//  4. Reads the response body using a bufio.Scanner to process it line-by-line.
//  5. For each non-empty line, applies cfg.Extractor to extract subdomain matches.
//  6. Streams each valid subdomain as a sources.Result of type ResultSubdomain.
//  7. Checks for any scanning errors; if found, streams an error result.
//  8. Closes the response body and the results channel upon completion.
func (source *Source) Run(domain string, cfg *sources.Configuration) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		getCertificateDetailsReqURL := fmt.Sprintf("https://certificatedetails.com/%s", domain)

		getCertificateDetailsRes, err := hqgohttp.Get(getCertificateDetailsReqURL)
		if err != nil && getCertificateDetailsRes.StatusCode != status.NotFound.Int() {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		scanner := bufio.NewScanner(getCertificateDetailsRes.Body)

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

			getCertificateDetailsRes.Body.Close()

			return
		}

		getCertificateDetailsRes.Body.Close()
	}()

	return results
}

// Name returns the unique identifier for the CertificateDetails data source.
// This identifier is used for logging, debugging, and to associate results
// with the correct data source.
//
// Returns:
//   - name (string): The constant sources.CERTIFICATEDETAILS representing the CertificateDetails source.
func (source *Source) Name() (name string) {
	return sources.CERTIFICATEDETAILS
}
