// Package securitytrails provides an implementation of the sources.Source interface
// for interacting with the SecurityTrails API.
//
// The SecurityTrails API offers comprehensive domain data, including subdomain information.
// This package defines a Source type that implements the Run and Name methods as specified
// by the sources.Source interface. The Run method sends a query to the SecurityTrails API,
// processes the JSON response, extracts subdomains, and streams discovered subdomains or errors
// via a channel.
package securitytrails

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	hqgohttp "go.source.hueristiq.com/http"
	"go.source.hueristiq.com/http/header"
	"go.source.hueristiq.com/http/mime"
)

// getSubdomainsResponse represents the structure of the JSON response returned by the SecurityTrails API.
//
// It contains the following fields:
//   - Endpoint: The API endpoint that processed the request.
//   - Meta: An object containing metadata about the request, including the scroll ID for pagination
//     and a flag indicating if the result limit was reached.
//   - Records: A slice of record objects, each containing a hostname.
//   - SubdomainCount: A boolean flag indicating if subdomain count data is available.
//   - Subdomains: A slice of strings representing the discovered subdomains.
type getSubdomainsResponse struct {
	Endpoint string `json:"endpoint"`
	Meta     struct {
		ScrollID     string `json:"scroll_id"`
		LimitReached string `json:"limit_reached"`
	} `json:"meta"`
	Records []struct {
		Hostname string `json:"hostname"`
	} `json:"records"`
	SubdomainCount bool     `json:"subdomain_count"`
	Subdomains     []string `json:"subdomains"`
}

// Source represents the SecurityTrails data source implementation.
// It implements the sources.Source interface, providing functionality
// for retrieving subdomains from the SecurityTrails API.
type Source struct{}

// Run initiates the process of retrieving subdomain information from the SecurityTrails API for a given domain.
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

		key, err := cfg.Keys.SecurityTrails.PickRandom()
		if key == "" || err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		getSubdomainsReqURL := fmt.Sprintf(
			"https://api.securitytrails.com/v1/domain/%s/subdomains",
			domain,
		)
		getSubdomainsReqCFG := &hqgohttp.RequestConfiguration{
			Params: map[string]string{
				"children_only":    "false",
				"include_inactive": "true",
			},
			Headers: map[string]string{
				header.Accept.String(): mime.JSON.String(),
				"APIKEY":               key,
			},
		}

		getSubdomainsRes, err := hqgohttp.Get(getSubdomainsReqURL, getSubdomainsReqCFG)
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
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
		}

		getSubdomainsRes.Body.Close()

		for _, subdomain := range getSubdomainsResData.Subdomains {
			if strings.HasSuffix(subdomain, ".") {
				subdomain += domain
			} else {
				subdomain = subdomain + "." + domain
			}

			result := sources.Result{
				Type:   sources.ResultSubdomain,
				Source: source.Name(),
				Value:  subdomain,
			}

			results <- result
		}
	}()

	return results
}

// Name returns the unique identifier for the data source.
// This identifier is used for logging, debugging, and associating results with the correct data source.
//
// Returns:
//   - name (string): The unique identifier for the data source.
func (source *Source) Name() string {
	return sources.SECURITYTRAILS
}
