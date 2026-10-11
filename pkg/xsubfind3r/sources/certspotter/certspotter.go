package certspotter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	hqgohttpheader "github.com/hueristiq/hq-lib-http-go/header"

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

func (s *Source) Run(ctx context.Context, cfg *sources.Configuration, domain string) <-chan sources.Result {
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

		getCTLogsSearchRes, err := cfg.Client.NewRequest().
			URL(getCTLogsSearchReqURL).
			Param("domain", domain).
			Param("include_subdomains", "true").
			Param("expand", "dns_names").
			Do(ctx)
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

			sources.CloseBody(getCTLogsSearchRes.Body)

			return
		}

		sources.CloseBody(getCTLogsSearchRes.Body)

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

			getCTLogsSearchRes, err := cfg.Client.NewRequest().
				URL(getCTLogsSearchReqURL).
				Param("domain", domain).
				Param("include_subdomains", "true").
				Param("expand", "dns_names").
				Param("after", id).
				SetHeader(hqgohttpheader.Authorization.String(), "Bearer "+key).
				Do(ctx)
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

				sources.CloseBody(getCTLogsSearchRes.Body)

				break
			}

			sources.CloseBody(getCTLogsSearchRes.Body)

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

func (s *Source) NeedsKeys() (needs bool) {
	needs = true

	return
}

func (s *Source) UseKeys(keys ...string) {
	s.keys = append(s.keys, keys...)
}

var _ sources.Source = (*Source)(nil)

func New() (source sources.Source) {
	source = &Source{
		keys: make(sources.Keys, 0),
	}

	return
}
