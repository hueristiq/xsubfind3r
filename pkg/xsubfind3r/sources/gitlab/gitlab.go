package gitlab

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/url"

	hqgohttpstatus "github.com/hueristiq/hq-lib-http-go/status"
	"github.com/spf13/cast"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type searchResponseItem struct {
	Data      string `json:"data"`
	ProjectID int    `json:"project_id"`
	Path      string `json:"path"`
	Ref       string `json:"ref"`
}

type Source struct {
	keys sources.Keys
}

func (s *Source) Name() (name string) {
	name = sources.Gitlab

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

		page := 1

		for page > 0 {
			searchReqURL := "https://gitlab.com/api/v4/search"

			searchRes, err := cfg.Client.NewRequest().
				URL(searchReqURL).
				Param("scope", "blobs").
				Param("search", domain).
				Param("per_page", "100").
				Param("page", cast.ToString(page)).
				SetHeader("PRIVATE-TOKEN", key).
				Do(ctx)
			if err != nil {
				if searchRes != nil {
					_ = searchRes.Body.Close()
				}

				result := sources.ErrorResult(s.Name(), fmt.Errorf("request failed: %w", err))

				results <- result

				return
			}

			if err = sources.CheckResponse(searchRes, key); err != nil {
				result := sources.ErrorResult(s.Name(), err)

				results <- result

				_ = searchRes.Body.Close()

				return
			}

			nextPage := searchRes.Header.Get("X-Next-Page")

			var items []searchResponseItem

			if err = json.NewDecoder(searchRes.Body).Decode(&items); err != nil {
				result := sources.ErrorResult(s.Name(), fmt.Errorf("failed to parse JSON response: %w", err))

				results <- result

				_ = searchRes.Body.Close()

				return
			}

			_ = searchRes.Body.Close()

			for _, item := range items {
				getRawContentReqURL := fmt.Sprintf(
					"https://gitlab.com/api/v4/projects/%d/repository/files/%s/raw?%s",
					item.ProjectID,
					url.PathEscape(item.Path),
					url.Values{"ref": {item.Ref}}.Encode(),
				)

				getRawContentRes, err := cfg.Client.NewRequest().
					URL(getRawContentReqURL).
					SetHeader("PRIVATE-TOKEN", key).
					Do(ctx)
				if err != nil {
					if getRawContentRes != nil {
						_ = getRawContentRes.Body.Close()
					}

					result := sources.ErrorResult(s.Name(), fmt.Errorf("request failed: %w", err))

					results <- result

					continue
				}

				if getRawContentRes.StatusCode != hqgohttpstatus.OK.Int() {
					_ = getRawContentRes.Body.Close()

					continue
				}

				scanner := bufio.NewScanner(getRawContentRes.Body)

				scanner.Buffer(make([]byte, 0, 64*1024), maxScanTokenSize)

				for scanner.Scan() {
					line := scanner.Text()

					if line == "" {
						continue
					}

					for _, candidate := range sources.ExtractSubdomains(cfg.Extractor, line) {
						subdomain, valid := cfg.Validator(candidate)
						if !valid {
							continue
						}

						result := sources.SubdomainResult(s.Name(), subdomain)

						results <- result
					}
				}

				if err = scanner.Err(); err != nil {
					result := sources.ErrorResult(s.Name(), fmt.Errorf("failed to read response body: %w", err))

					results <- result

					_ = getRawContentRes.Body.Close()

					return
				}

				_ = getRawContentRes.Body.Close()
			}

			page = cast.ToInt(nextPage)
		}
	}()

	return results
}

const maxScanTokenSize = 16 << 20

var _ sources.Source = (*Source)(nil)

func New() (source sources.Source) {
	source = &Source{
		keys: make(sources.Keys, 0),
	}

	return
}
