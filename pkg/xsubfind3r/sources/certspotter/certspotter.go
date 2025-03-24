package certspotter

import (
	"encoding/json"
	"strings"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	hqgohttp "go.source.hueristiq.com/http"
)

type getCTLogsSearchResponse struct {
	ID       string   `json:"id"`
	DNSNames []string `json:"dns_names"`
}

type Source struct{}

func (source *Source) Run(cfg *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		key, err := cfg.Keys.Certspotter.PickRandom()
		if key == "" || err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
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
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var getCTLogsSearchResData []getCTLogsSearchResponse

		if err = json.NewDecoder(getCTLogsSearchRes.Body).Decode(&getCTLogsSearchResData); err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
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
					Source: source.Name(),
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
				Headers: map[string]string{
					"Authorization": "Bearer " + key,
				},
			}

			getCTLogsSearchRes, err := hqgohttp.Get(getCTLogsSearchReqURL, getCTLogsSearchReqCFG)
			if err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: source.Name(),
					Error:  err,
				}

				results <- result

				break
			}

			var getCTLogsSearchResData []getCTLogsSearchResponse

			if err = json.NewDecoder(getCTLogsSearchRes.Body).Decode(&getCTLogsSearchResData); err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: source.Name(),
					Error:  err,
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
						Source: source.Name(),
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

func (source *Source) Name() string {
	return sources.CERTSPOTTER
}
