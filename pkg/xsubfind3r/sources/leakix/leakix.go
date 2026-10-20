package leakix

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	hqgohttpheader "github.com/hueristiq/hq-lib-http-go/header"
	hqgohttpmime "github.com/hueristiq/hq-lib-http-go/mime"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type getSubdomainsResponse struct {
	Subdomain   string    `json:"subdomain"`
	DistinctIps int       `json:"distinct_ips"`
	LastSeen    time.Time `json:"last_seen"`
}

type Source struct {
	keys sources.Keys
}

func (s *Source) Name() (name string) {
	name = sources.Leakix

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
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  fmt.Errorf("failed to select key: %w", err),
			}

			results <- result

			return
		}

		getSubdomainsReqURL := "https://leakix.net/api/subdomains/" + domain

		getSubdomainsRes, err := cfg.Client.NewRequest().
			URL(getSubdomainsReqURL).
			SetHeader(hqgohttpheader.Accept.String(), hqgohttpmime.JSON.String()).
			SetHeader("api-key", key).
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

		var getSubdomainsResData []getSubdomainsResponse

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

		for _, record := range getSubdomainsResData {
			subdomain, valid := cfg.Validator(record.Subdomain)
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
