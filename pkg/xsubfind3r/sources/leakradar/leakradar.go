package leakradar

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cast"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type getURLsResponse struct {
	Items    []item `json:"items"`
	Total    int    `json:"total"`
	Page     int    `json:"page"`
	PageSize int    `json:"page_size"`
}

type item struct {
	URL         string `json:"url"`
	Occurrences int    `json:"occurrences"`
}

type Source struct {
	keys sources.Keys
}

func (s *Source) Name() (name string) {
	name = sources.Leakradar

	return
}

func (s *Source) NeedsKeys() (needs bool) {
	needs = true

	return
}

func (s *Source) UseKeys(keys ...string) {
	s.keys = append(s.keys, keys...)
}

func (s *Source) Run(ctx context.Context, cfg *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		for page := 1; ; page++ {
			key, err := s.keys.PickRandom()
			if err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: s.Name(),
					Error:  fmt.Errorf("failed to select key: %w", err),
				}

				results <- result

				return
			}

			getURLsReqURL := fmt.Sprintf("https://api.leakradar.io/search/domain/%s/urls", domain)

			getURLsRes, err := cfg.Client.NewRequest().
				URL(getURLsReqURL).
				Param("page", cast.ToString(page)).
				Param("page_size", "100").
				SetHeader("Authorization", "Bearer "+key).
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

			var getURLsResData getURLsResponse

			if err = json.NewDecoder(getURLsRes.Body).Decode(&getURLsResData); err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: s.Name(),
					Error:  fmt.Errorf("failed to parse JSON response: %w", err),
				}

				results <- result

				_ = getURLsRes.Body.Close()

				return
			}

			_ = getURLsRes.Body.Close()

			for _, item := range getURLsResData.Items {
				URL := item.URL

				for _, candidate := range sources.ExtractSubdomains(cfg.Extractor, URL) {
					subdomain, valid := cfg.Validator(candidate)
					if !valid {
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

			if getURLsResData.PageSize*getURLsResData.Page >= getURLsResData.Total {
				break
			}
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
