package threatminer

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type getSubdomainsResponse struct {
	StatusCode    string   `json:"status_code"`
	StatusMessage string   `json:"status_message"`
	Results       []string `json:"results"`
}

type Source struct{}

func (s *Source) Name() (name string) {
	name = sources.Threatminer

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

		getSubdomainsReqURL := "https://api.threatminer.org/v2/domain.php"

		getSubdomainsRes, err := cfg.Client.NewRequest().
			URL(getSubdomainsReqURL).
			Param("q", domain).
			Param("rt", "5").
			Do(ctx)
		if err != nil {
			result := sources.ErrorResult(s.Name(), fmt.Errorf("request failed: %w", err))

			results <- result

			return
		}

		if err = sources.CheckResponse(getSubdomainsRes); err != nil {
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

		for _, subdomain := range getSubdomainsResData.Results {
			subdomain, valid := cfg.Validator(subdomain)
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
	source = &Source{}

	return
}
