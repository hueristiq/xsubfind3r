package merklemap

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/spf13/cast"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type getSubdomainsResponse struct {
	Count   int `json:"count"`
	Results []struct {
		Hostname string `json:"hostname"`
	} `json:"results"`
}

type Source struct {
	keys sources.Keys
}

func (s *Source) Name() (name string) {
	name = sources.Merklemap

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

		key, err := s.keys.PickRandom()
		if err != nil {
			result := sources.ErrorResult(s.Name(), fmt.Errorf("failed to select key: %w", err))

			results <- result

			return
		}

		processed := 0
		total := 0
		lastCount := -1

		page := 0

		for ; page < sources.MaxPages; page++ {
			getSubdomainsReqURL := "https://api.merklemap.com/v1/search"

			getSubdomainsRes, err := cfg.Client.Request().
				URL(getSubdomainsReqURL).
				Param("query", "*."+domain).
				Param("page", cast.ToString(page)).
				SetHeader("Accept", "application/json").
				SetHeader("Authorization", "Bearer "+key).
				Context(ctx).Do()
			if err != nil {
				result := sources.ErrorResult(s.Name(), fmt.Errorf("request failed: %w", err))

				results <- result

				return
			}

			if err = sources.CheckResponse(getSubdomainsRes.Std(), key); err != nil {
				result := sources.ErrorResult(s.Name(), err)

				results <- result

				_ = getSubdomainsRes.Body.Close()

				return
			}

			var getSubdomainsResData getSubdomainsResponse

			if err = json.NewDecoder(getSubdomainsRes.Body).Decode(&getSubdomainsResData); err != nil {
				result := sources.ErrorResult(s.Name(), fmt.Errorf("failed to parse JSON response: %w", err))

				results <- result

				_ = getSubdomainsRes.Body.Close()

				return
			}

			_ = getSubdomainsRes.Body.Close()

			if page == 0 {
				total = getSubdomainsResData.Count
			}

			if len(getSubdomainsResData.Results) == 0 {
				break
			}

			before := processed

			for _, record := range getSubdomainsResData.Results {
				subdomain, valid := cfg.Validator(record.Hostname)
				if !valid {
					continue
				}

				result := sources.SubdomainResult(s.Name(), subdomain)

				results <- result

				processed++
			}

			// A repeated total with no progress means the server is
			// serving the same page again.
			if page > 0 && getSubdomainsResData.Count == lastCount && processed == before {
				break
			}

			lastCount = getSubdomainsResData.Count

			if processed >= total {
				break
			}
		}

		if page == sources.MaxPages {
			result := sources.WarningResult(s.Name(), fmt.Sprintf("pagination capped at %d pages", sources.MaxPages))

			results <- result
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
