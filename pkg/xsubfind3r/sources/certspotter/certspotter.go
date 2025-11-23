package certspotter

import (
	"encoding/json"
	"fmt"
	"strings"

	hqgohttp "github.com/hueristiq/hq-go-http"
	hqgohttpheader "github.com/hueristiq/hq-go-http/header"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type getCTLogsSearchResponse struct {
	ID       string   `json:"id"`
	DNSNames []string `json:"dns_names"`
}

type Source struct {
	keys sources.Keys
}

func (s *Source) Name() (name string) {
	name = sources.CERTSPOTTER

	return
}

func (s *Source) UseKeys(keys ...string) {
	s.keys = append(s.keys, keys...)
}

func (s *Source) Run(cfg *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		key, err := s.keys.PickRandom()
		if key == "" || err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  fmt.Errorf("failed to select key: %w", err),
			}

			results <- result

			return
		}

		getCTLogsSearchReqURL := "https://api.certspotter.com/v1/issuances"
		getCTLogsSearchReqCFG := &hqgohttp.RequestConfiguration{
			Params: map[string]string{
				"domain":             domain,
				"include_subdomains": "true",
				"expand":             "dns_names",
			},
		}

		getCTLogsSearchRes, err := hqgohttp.Get(getCTLogsSearchReqURL, getCTLogsSearchReqCFG)
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  fmt.Errorf("request failed: %w", err),
			}

			results <- result

			return
		}

		var getCTLogsSearchResData []getCTLogsSearchResponse

		if err = json.NewDecoder(getCTLogsSearchRes.Body).Decode(&getCTLogsSearchResData); err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  fmt.Errorf("failed to parse JSON response: %w", err),
			}

			results <- result

			getCTLogsSearchRes.Body.Close()

			return
		}

		getCTLogsSearchRes.Body.Close()

		if len(getCTLogsSearchResData) == 0 {
			return
		}

		for _, cert := range getCTLogsSearchResData {
			for _, subdomain := range cert.DNSNames {
				if subdomain != domain && !strings.HasSuffix(subdomain, "."+domain) {
					continue
				}

				result := sources.Result{
					Type:   sources.ResultSubdomain,
					Source: s.Name(),
					Value:  subdomain,
				}

				results <- result
			}
		}

		id := getCTLogsSearchResData[len(getCTLogsSearchResData)-1].ID

		for {
			getCTLogsSearchReqURL := "https://api.certspotter.com/v1/issuances"
			getCTLogsSearchReqCFG := &hqgohttp.RequestConfiguration{
				Params: map[string]string{
					"domain":             domain,
					"include_subdomains": "true",
					"expand":             "dns_names",
					"after":              id,
				},
				Headers: []hqgohttp.Header{
					hqgohttp.NewSetHeader(hqgohttpheader.Authorization.String(), "Bearer "+key),
				},
			}

			getCTLogsSearchRes, err := hqgohttp.Get(getCTLogsSearchReqURL, getCTLogsSearchReqCFG)
			if err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: s.Name(),
					Error:  fmt.Errorf("request failed: %w", err),
				}

				results <- result

				break
			}

			var getCTLogsSearchResData []getCTLogsSearchResponse

			if err = json.NewDecoder(getCTLogsSearchRes.Body).Decode(&getCTLogsSearchResData); err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: s.Name(),
					Error:  fmt.Errorf("failed to parse JSON response: %w", err),
				}

				results <- result

				getCTLogsSearchRes.Body.Close()

				break
			}

			getCTLogsSearchRes.Body.Close()

			if len(getCTLogsSearchResData) == 0 {
				break
			}

			for _, cert := range getCTLogsSearchResData {
				for _, subdomain := range cert.DNSNames {
					if subdomain != domain && !strings.HasSuffix(subdomain, "."+domain) {
						continue
					}

					result := sources.Result{
						Type:   sources.ResultSubdomain,
						Source: s.Name(),
						Value:  subdomain,
					}

					results <- result
				}
			}

			id = getCTLogsSearchResData[len(getCTLogsSearchResData)-1].ID
		}
	}()

	return results
}

var _ sources.Source = (*Source)(nil)

func New() (source sources.Source) {
	source = &Source{
		keys: make(sources.Keys, 0),
	}

	return
}
