package dnsdb

import (
	"bufio"
	"bytes"
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

const (
	baseReqURL = "https://api.dnsdb.info/dnsdb/v2"
	jsonNull   = "null"
)

type rateLimitResponse struct {
	Rate struct {
		OffsetMax json.RawMessage `json:"offset_max"`
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

		offsetMax, err := getMaxOffset(ctx, cfg, key)
		if err != nil {
			result := sources.ErrorResult(s.Name(), fmt.Errorf("failed to get rate limit: %w", err))

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
				Do(ctx) //nolint:bodyclose // closed by the reader goroutine and every cancellation path below.
			if err != nil {
				result := sources.ErrorResult(s.Name(), fmt.Errorf("request failed: %w", err))

				results <- result

				return
			}

			if err = sources.CheckResponse(getSubdomainsRes, key); err != nil {
				result := sources.ErrorResult(s.Name(), err)

				results <- result

				_ = getSubdomainsRes.Body.Close()

				return
			}

			var condition string

			aborted := false

			done := make(chan struct{})

			go func() {
				defer close(done)

				reader := bufio.NewReader(getSubdomainsRes.Body)

				processLine := func(line []byte) (stop bool) {
					line = bytes.TrimSpace(line)

					if len(line) == 0 {
						return false
					}

					var record lookupResponse

					if parseErr := json.Unmarshal(line, &record); parseErr != nil {
						result := sources.ErrorResult(s.Name(), fmt.Errorf("failed to parse JSON response: %w", parseErr))

						results <- result

						aborted = true

						return true
					}

					condition = record.Condition

					if condition != "" && condition != "ongoing" && condition != "begin" {
						return true
					}

					if record.Obj.Name != "" {
						subdomain, valid := cfg.Validator(record.Obj.Name)
						if !valid {
							return false
						}

						result := sources.SubdomainResult(s.Name(), subdomain)

						results <- result

						emitted++
					}

					return false
				}

				for {
					line, readErr := reader.ReadBytes('\n')

					// The final record may not be newline-terminated.
					if len(line) > 0 && processLine(line) {
						return
					}

					if readErr == nil {
						continue
					}

					if errors.Is(readErr, io.EOF) {
						return
					}

					if ctx.Err() == nil {
						result := sources.ErrorResult(s.Name(), fmt.Errorf("failed to read response body: %w", readErr))

						results <- result

						aborted = true
					}

					return
				}
			}()

			select {
			case <-done:
			case <-ctx.Done():
				_ = getSubdomainsRes.Body.Close()

				<-done
			}

			if ctx.Err() != nil || aborted {
				_ = getSubdomainsRes.Body.Close()

				return
			}

			_ = getSubdomainsRes.Body.Close()

			if condition == "limited" {
				if emitted > before && offsetMax != 0 && emitted <= offsetMax {
					params.Set("offset", strconv.FormatUint(emitted, 10))

					continue
				}

				result := sources.ErrorResult(s.Name(), errors.New("lookup truncated: pagination limit reached"))

				results <- result
			} else if condition != "succeeded" {
				result := sources.ErrorResult(s.Name(), fmt.Errorf("lookup terminated with condition: %s", condition))

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

	defer func() {
		_ = getRateLimitRes.Body.Close()
	}()

	if err = sources.CheckResponse(getRateLimitRes, key); err != nil {
		return
	}

	var getRateLimitResData rateLimitResponse

	if err = json.NewDecoder(getRateLimitRes.Body).Decode(&getRateLimitResData); err != nil {
		return
	}

	offsetMax = parseOffsetMax(getRateLimitResData.Rate.OffsetMax)

	return
}

// parseOffsetMax tolerates the API returning the offset as a number, a
// numeric string, "n/a", or omitting it entirely. Anything that cannot be
// parsed is treated as unlimited (zero).
func parseOffsetMax(raw json.RawMessage) (offsetMax uint64) {
	value := strings.TrimSpace(string(raw))

	if value == "" || value == jsonNull {
		return
	}

	if value[0] == '"' {
		var text string

		if err := json.Unmarshal(raw, &text); err != nil {
			return
		}

		value = strings.TrimSpace(text)
	}

	if strings.EqualFold(value, "n/a") {
		return
	}

	offsetMax, _ = strconv.ParseUint(value, 10, 64)

	return
}

var _ sources.Source = (*Source)(nil)

func New() (source sources.Source) {
	source = &Source{
		keys: make(sources.Keys, 0),
	}

	return
}
