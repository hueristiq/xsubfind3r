package c99

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type getSubdomainsResponse struct {
	Success    bool `json:"success"`
	Subdomains []struct {
		Subdomain string `json:"subdomain"`
	} `json:"subdomains"`
	Error string `json:"error"`
}

type Source struct {
	keys sources.Keys
}

func (s *Source) Name() (name string) {
	name = sources.C99

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

		getSubdomainsReqURL := fmt.Sprintf("https://api.c99.nl/subdomainfinder?key=%s&domain=%s&json", key, domain)

		getSubdomainsRes, err := cfg.Client.NewRequest().
			URL(getSubdomainsReqURL).
			Do(ctx)
		if err != nil {
			result := sources.ErrorResult(s.Name(), fmt.Errorf("request failed: %w", sources.RedactError(err, key)))

			results <- result

			return
		}

		if err = sources.CheckResponse(getSubdomainsRes, key); err != nil {
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

		if getSubdomainsResData.Error != "" {
			result := sources.ErrorResult(s.Name(), fmt.Errorf("%s", getSubdomainsResData.Error))

			results <- result

			return
		}

		for _, record := range getSubdomainsResData.Subdomains {
			subdomain, valid := cfg.Validator(record.Subdomain)
			if !valid {
				continue
			}

			result := sources.SubdomainResult(s.Name(), subdomain)

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
