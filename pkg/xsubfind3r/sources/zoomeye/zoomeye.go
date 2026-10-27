package zoomeye

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"

	hqgohttpmethod "github.com/hueristiq/hq-lib-http-go/method"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type searchRequest struct {
	Query    string `json:"qbase64"`
	Page     uint   `json:"page"`
	PageSize uint   `json:"pagesize"`
	SubType  string `json:"sub_type"`
}

type searchResponse struct {
	Total int `json:"total"`
	Data  []struct {
		URL    string `json:"url"`
		Domain string `json:"domain"`
	} `json:"data"`
}

type Source struct {
	keys sources.Keys
}

func (s *Source) Name() (name string) {
	name = sources.Zoomeye

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

		getSubdomainsReqURL := "https://api.zoomeye.ai/v2/search"

		processed := 0

		for page := uint(1); ; page++ {
			searchReqBody, err := json.Marshal(searchRequest{
				Query:    base64.StdEncoding.EncodeToString([]byte("*." + domain)),
				Page:     page,
				PageSize: 100,
				SubType:  "web",
			})
			if err != nil {
				result := sources.ErrorResult(s.Name(), fmt.Errorf("failed to marshal search request: %w", err))

				results <- result

				return
			}

			getSubdomainsRes, err := cfg.Client.Request().
				Method(hqgohttpmethod.POST).
				URL(getSubdomainsReqURL).
				SetHeader("API-KEY", key).
				SetHeader("Content-Type", "application/json").
				Body(searchReqBody).
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

			var getSubdomainsResData searchResponse

			if err = json.NewDecoder(getSubdomainsRes.Body).Decode(&getSubdomainsResData); err != nil {
				result := sources.ErrorResult(s.Name(), fmt.Errorf("failed to parse JSON response: %w", err))

				results <- result

				_ = getSubdomainsRes.Body.Close()

				return
			}

			_ = getSubdomainsRes.Body.Close()

			if len(getSubdomainsResData.Data) == 0 {
				break
			}

			for _, record := range getSubdomainsResData.Data {
				for _, target := range []string{record.Domain, record.URL} {
					for _, candidate := range sources.ExtractSubdomains(cfg.Extractor, target) {
						subdomain, valid := cfg.Validator(candidate)
						if !valid {
							continue
						}

						result := sources.SubdomainResult(s.Name(), subdomain)

						results <- result
					}
				}

				processed++
			}

			if processed >= getSubdomainsResData.Total {
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
