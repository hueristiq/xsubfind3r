package shodan

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type getDNSResponse struct {
	Domain     string   `json:"domain"`
	Subdomains []string `json:"subdomains"`
	Result     int      `json:"result"`
	Error      string   `json:"error"`
}

type Source struct {
	keys sources.Keys
}

func (s *Source) Name() (name string) {
	name = sources.Shodan

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
		if key == "" || err != nil {
			result := sources.ErrorResult(s.Name(), fmt.Errorf("failed to select key: %w", err))

			results <- result

			return
		}

		getDNSReqURL := "https://api.shodan.io/dns/domain/" + domain

		getDNSRes, err := cfg.Client.NewRequest().
			URL(getDNSReqURL).
			Param("key", key).
			Do(ctx)
		if err != nil {
			result := sources.ErrorResult(s.Name(), fmt.Errorf("request failed: %w", sources.RedactError(err, key)))

			results <- result

			return
		}

		if err = sources.CheckResponse(getDNSRes, key); err != nil {
			result := sources.ErrorResult(s.Name(), err)

			results <- result

			_ = getDNSRes.Body.Close()

			return
		}

		var getDNSResData getDNSResponse

		if err = json.NewDecoder(getDNSRes.Body).Decode(&getDNSResData); err != nil {
			result := sources.ErrorResult(s.Name(), fmt.Errorf("failed to parse JSON response: %w", err))

			results <- result

			_ = getDNSRes.Body.Close()

			return
		}

		_ = getDNSRes.Body.Close()

		for _, subdomain := range getDNSResData.Subdomains {
			subdomain, valid := cfg.Validator(fmt.Sprintf("%s.%s", subdomain, domain))
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
