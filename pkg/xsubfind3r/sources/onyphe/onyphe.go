package onyphe

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cast"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type result struct {
	Subdomains []string
	Hostname   string
	Forward    string
	Reverse    string
}

func decodeStringOrSlice(raw json.RawMessage) (values []string, err error) {
	if len(raw) == 0 {
		return
	}

	if err = json.Unmarshal(raw, &values); err != nil {
		var single string

		if err = json.Unmarshal(raw, &single); err != nil {
			return
		}

		values = []string{single}
	}

	return
}

func (r *result) UnmarshalJSON(data []byte) (err error) {
	var raw struct {
		Subdomains json.RawMessage `json:"subdomains"`
		Hostname   json.RawMessage `json:"hostname"`
		Forward    json.RawMessage `json:"forward"`
		Reverse    json.RawMessage `json:"reverse"`
	}

	if err = json.Unmarshal(data, &raw); err != nil {
		return
	}

	if r.Subdomains, err = decodeStringOrSlice(raw.Subdomains); err != nil {
		return
	}

	if len(raw.Hostname) > 0 {
		var values []string

		if values, err = decodeStringOrSlice(raw.Hostname); err != nil {
			return
		}

		if len(values) > 0 {
			r.Hostname = values[0]
		}
	}

	_ = json.Unmarshal(raw.Forward, &r.Forward)
	_ = json.Unmarshal(raw.Reverse, &r.Reverse)

	return
}

type getSubdomainsResponse struct {
	Results []result        `json:"results"`
	MaxPage json.RawMessage `json:"max_page"`
}

type Source struct {
	keys sources.Keys
}

func (s *Source) Name() (name string) {
	name = sources.ONYPHE

	return
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

		page := 1
		maxPage := 1

		for {
			getSubdomainsReqURL := "https://www.onyphe.io/api/v2/search/"

			getSubdomainsRes, err := cfg.Client.NewRequest().
				URL(getSubdomainsReqURL).
				Param("q", "category:resolver domain:"+domain).
				Param("page", cast.ToString(page)).
				Param("size", "1000").
				SetHeader("Authorization", "bearer "+key).
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

				sources.CloseBody(getSubdomainsRes.Body)

				return
			}

			sources.CloseBody(getSubdomainsRes.Body)

			if page == 1 {
				maxPage = cast.ToInt(strings.Trim(string(getSubdomainsResData.MaxPage), `"`))

				if maxPage < 1 {
					maxPage = 1
				}
			}

			if len(getSubdomainsResData.Results) == 0 {
				break
			}

			for _, record := range getSubdomainsResData.Results {
				for _, target := range append(record.Subdomains, record.Hostname, record.Forward, record.Reverse) {
					subdomains := cfg.Extractor.FindAllString(target, -1)

					for _, subdomain := range subdomains {
						result := sources.Result{
							Type:   sources.ResultSubdomain,
							Source: s.Name(),
							Value:  subdomain,
						}

						results <- result
					}
				}
			}

			if page >= maxPage {
				break
			}

			page++
		}
	}()

	return results
}

func (s *Source) NeedsKeys() (needs bool) {
	needs = true

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
