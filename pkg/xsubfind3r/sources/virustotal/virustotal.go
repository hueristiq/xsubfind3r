package virustotal

import (
	"context"
	"encoding/json"
	"fmt"

	hqgolimiter "github.com/hueristiq/hq-lib-limiter-go"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type getSubdomainsResponse struct {
	Error struct {
		Code    string `json:"code"`
		Message string `json:"message"`
	} `json:"error"`
	Data []struct {
		ID    string `json:"id"`
		Type  string `json:"type"`
		Links struct {
			Self string `json:"self"`
		} `json:"links"`
	} `json:"data"`
	Meta struct {
		Cursor string `json:"cursor"`
	} `json:"meta"`
}

type Source struct {
	keys sources.Keys
}

func (s *Source) Name() (name string) {
	name = sources.Virustotal

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
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  fmt.Errorf("failed to select key: %w", err),
			}

			results <- result

			return
		}

		var cursor string

		for {
			if err := limiter.Wait(ctx); err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: s.Name(),
					Error:  fmt.Errorf("failed to wait on rate limiter: %w", err),
				}

				results <- result

				break
			}

			getSubdomainsReqURL := fmt.Sprintf("https://www.virustotal.com/api/v3/domains/%s/subdomains", domain)

			getSubdomainsReq := cfg.Client.NewRequest().
				URL(getSubdomainsReqURL).
				Param("limit", "40").
				SetHeader("x-apikey", key)

			if cursor != "" {
				getSubdomainsReq = getSubdomainsReq.Param("cursor", cursor)
			}

			getSubdomainsRes, err := getSubdomainsReq.Do(ctx)
			if err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: s.Name(),
					Error:  fmt.Errorf("request failed: %w", sources.RedactError(err, key)),
				}

				results <- result

				break
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

				break
			}

			_ = getSubdomainsRes.Body.Close()

			if getSubdomainsResData.Error.Message != "" {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: s.Name(),
					Error:  fmt.Errorf("domain error: %s, %s", getSubdomainsResData.Error.Code, getSubdomainsResData.Error.Message),
				}

				results <- result

				break
			}

			for _, record := range getSubdomainsResData.Data {
				subdomain := record.ID

				result := sources.Result{
					Type:   sources.ResultSubdomain,
					Source: s.Name(),
					Value:  subdomain,
				}

				results <- result
			}

			cursor = getSubdomainsResData.Meta.Cursor

			if cursor == "" {
				break
			}
		}
	}()

	return results
}

var limiter = hqgolimiter.New(
	hqgolimiter.WithRequestsPerMinute(4),
)

var _ sources.Source = (*Source)(nil)

func New() (source sources.Source) {
	source = &Source{
		keys: make(sources.Keys, 0),
	}

	return
}
