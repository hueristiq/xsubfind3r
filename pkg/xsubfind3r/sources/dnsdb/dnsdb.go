package dnsdb

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/url"
	"strconv"
	"strings"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

const baseReqURL = "https://api.dnsdb.info/dnsdb/v2"

type rateLimitResponse struct {
	Rate struct {
		OffsetMax json.Number `json:"offset_max"`
	} `json:"rate"`
}

type lookupResponse struct {
	Condition string `json:"cond"`
	Obj       struct {
		Name string `json:"rrname"`
	} `json:"obj"`
}

type Source struct {
	keys sources.Keys
}

func (s *Source) Name() (name string) {
	name = sources.DNSDB

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

		offsetMax, err := getMaxOffset(ctx, cfg, key)
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  fmt.Errorf("failed to get rate limit: %w", err),
			}

			results <- result

			return
		}

		params := url.Values{}
		params.Add("limit", "0")
		// The API requires a client identifier; identify honestly.
		params.Add("swclient", "xsubfind3r")

		var emitted uint64

		for {
			before := emitted
			getSubdomainsReqURL := fmt.Sprintf("%s/lookup/rrset/name/*.%s?%s", baseReqURL, domain, params.Encode())

			getSubdomainsRes, err := cfg.Client.NewRequest().
				URL(getSubdomainsReqURL).
				SetHeader("X-API-KEY", key).
				SetHeader("Accept", "application/x-ndjson").
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

			var condition string

			reader := bufio.NewReader(getSubdomainsRes.Body)

			for {
				line, err := reader.ReadBytes('\n')
				if errors.Is(err, io.EOF) {
					break
				}

				if err != nil {
					result := sources.Result{
						Type:   sources.ResultError,
						Source: s.Name(),
						Error:  fmt.Errorf("failed to read response body: %w", err),
					}

					results <- result

					sources.CloseBody(getSubdomainsRes.Body)

					return
				}

				var record lookupResponse

				if err = json.Unmarshal(line, &record); err != nil {
					result := sources.Result{
						Type:   sources.ResultError,
						Source: s.Name(),
						Error:  fmt.Errorf("failed to parse JSON response: %w", err),
					}

					results <- result

					sources.CloseBody(getSubdomainsRes.Body)

					return
				}

				condition = record.Condition

				if condition == "" || condition == "ongoing" {
					if record.Obj.Name != "" {
						result := sources.Result{
							Type:   sources.ResultSubdomain,
							Source: s.Name(),
							Value:  strings.TrimSuffix(record.Obj.Name, "."),
						}

						results <- result

						emitted++
					}
				} else if condition != "begin" {
					break
				}
			}

			sources.CloseBody(getSubdomainsRes.Body)

			if condition == "limited" {
				if emitted > before && offsetMax != 0 && emitted <= offsetMax {
					params.Set("offset", strconv.FormatUint(emitted, 10))

					continue
				}
			} else if condition != "succeeded" {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: s.Name(),
					Error:  fmt.Errorf("lookup terminated with condition: %s", condition),
				}

				results <- result
			}

			break
		}
	}()

	return results
}

func getMaxOffset(ctx context.Context, cfg *sources.Configuration, key string) (offsetMax uint64, err error) {
	getRateLimitReqURL := baseReqURL + "/rate_limit"

	getRateLimitRes, err := cfg.Client.NewRequest().
		URL(getRateLimitReqURL).
		SetHeader("X-API-KEY", key).
		Do(ctx)
	if err != nil {
		return
	}

	defer sources.CloseBody(getRateLimitRes.Body)

	var getRateLimitResData rateLimitResponse

	if err = json.NewDecoder(getRateLimitRes.Body).Decode(&getRateLimitResData); err != nil {
		return
	}

	if getRateLimitResData.Rate.OffsetMax.String() != "n/a" {
		offsetMax, err = strconv.ParseUint(getRateLimitResData.Rate.OffsetMax.String(), 10, 64)
		if err != nil {
			return
		}
	}

	return
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
