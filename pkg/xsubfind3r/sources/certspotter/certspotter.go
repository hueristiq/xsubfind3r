package certspotter

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	hqgohttpheader "github.com/hueristiq/hq-lib-http-go/header"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

const maxPages = 100

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

		key, _ := s.keys.PickRandom()

		after := ""

		for range maxPages {
			getCTLogsSearchReqURL := "https://api.certspotter.com/v1/issuances"

			req := cfg.Client.NewRequest().
				URL(getCTLogsSearchReqURL).
				Param("domain", domain).
				Param("include_subdomains", "true").
				Param("expand", "dns_names")

			if key != "" {
				req = req.SetHeader(hqgohttpheader.Authorization.String(), "Bearer "+key)
			}

			if after != "" {
				req = req.Param("after", after)
			}

			getCTLogsSearchRes, err := req.Do(ctx)
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

			after = getCTLogsSearchResData[len(getCTLogsSearchResData)-1].ID
		}
	}()

	return results
}

func (s *Source) NeedsKeys() (needs bool) {
	needs = false

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
