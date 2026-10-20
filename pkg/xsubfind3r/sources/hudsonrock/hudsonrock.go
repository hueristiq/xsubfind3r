package hudsonrock

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type getSubdomainsResponse struct {
	Data struct {
		EmployeesUrls []struct {
			URL string `json:"url"`
		} `json:"employees_urls"`
		ClientsUrls []struct {
			URL string `json:"url"`
		} `json:"clients_urls"`
	} `json:"data"`
}

type Source struct{}

func (s *Source) Name() (name string) {
	name = sources.Hudsonrock

	return
}

func (s *Source) NeedsKeys() (needs bool) {
	needs = false

	return
}

func (s *Source) UseKeys(keys ...string) {
}

func (s *Source) Run(ctx context.Context, cfg *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		getSubdomainsReqURL := "https://cavalier.hudsonrock.com/api/json/v2/osint-tools/urls-by-domain"

		getSubdomainsRes, err := cfg.Client.NewRequest().
			URL(getSubdomainsReqURL).
			Param("domain", domain).
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

		var getSubdomainsResData getSubdomainsResponse

		if err = json.NewDecoder(getSubdomainsRes.Body).Decode(&getSubdomainsResData); err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  fmt.Errorf("failed to parse JSON response: %w", err),
			}

			results <- result

			_ = getSubdomainsRes.Body.Close()

			return
		}

		_ = getSubdomainsRes.Body.Close()

		for _, record := range append(getSubdomainsResData.Data.EmployeesUrls, getSubdomainsResData.Data.ClientsUrls...) {
			subdomains := cfg.Extract(record.URL)

			for _, subdomain := range subdomains {
				result := sources.Result{
					Type:   sources.ResultSubdomain,
					Source: s.Name(),
					Value:  subdomain,
				}

				results <- result
			}
		}
	}()

	return results
}

var _ sources.Source = (*Source)(nil)

func New() (source sources.Source) {
	source = &Source{}

	return
}
