package urlscan

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	hqgohttpheader "github.com/hueristiq/hq-lib-http-go/header"
	hqgohttpmime "github.com/hueristiq/hq-lib-http-go/mime"
	hqgohttpstatus "github.com/hueristiq/hq-lib-http-go/status"
	"github.com/spf13/cast"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type searchResponse struct {
	Results []struct {
		Page struct {
			Domain   string `json:"domain"`
			MimeType string `json:"mimeType"`
			URL      string `json:"url"`
			Status   string `json:"status"`
		} `json:"page"`
		Sort []interface{} `json:"sort"`
	} `json:"results"`
	Status  int  `json:"status"`
	Total   int  `json:"total"`
	Took    int  `json:"took"`
	HasMore bool `json:"has_more"`
}

type Source struct {
	keys sources.Keys
}

func (s *Source) Name() (name string) {
	name = sources.URLScan

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

		var after string

		page := 0

		for ; page < sources.MaxPages; page++ {
			searchReqURL := "https://urlscan.io/api/v1/search"

			searchReq := cfg.Client.Request().
				URL(searchReqURL).
				Param("q", "domain:"+domain).
				Param("size", "10000").
				SetHeader(hqgohttpheader.Accept.String(), hqgohttpmime.JSON.String()).
				SetHeader("API-Key", key)

			if after != "" {
				searchReq = searchReq.Param("search_after", after)
			}

			searchRes, err := searchReq.Context(ctx).Do()
			if err != nil {
				result := sources.ErrorResult(s.Name(), fmt.Errorf("request failed: %w", err))

				results <- result

				break
			}

			if err = sources.CheckResponse(searchRes.Std(), key); err != nil {
				result := sources.ErrorResult(s.Name(), err)

				results <- result

				_ = searchRes.Body.Close()

				break
			}

			var searchResData searchResponse

			if err = json.NewDecoder(searchRes.Body).Decode(&searchResData); err != nil {
				result := sources.ErrorResult(s.Name(), fmt.Errorf("failed to parse JSON response: %w", err))

				results <- result

				_ = searchRes.Body.Close()

				break
			}

			_ = searchRes.Body.Close()

			if searchResData.Status == hqgohttpstatus.TooManyRequests.Int() {
				result := sources.ErrorResult(s.Name(), fmt.Errorf("request failed with status: %v", searchResData.Status))

				results <- result

				break
			}

			for _, item := range searchResData.Results {
				subdomain, valid := cfg.Validator(item.Page.Domain)
				if !valid {
					continue
				}

				result := sources.SubdomainResult(s.Name(), subdomain)

				results <- result
			}

			if !searchResData.HasMore {
				break
			}

			if len(searchResData.Results) < 1 {
				break
			}

			lastResult := searchResData.Results[len(searchResData.Results)-1]

			// Without a sort cursor the next page would repeat this
			// one forever.
			if len(lastResult.Sort) < 1 {
				break
			}

			var sortValues []string

			for index := range lastResult.Sort {
				sortValues = append(sortValues, cast.ToString(lastResult.Sort[index]))
			}

			nextAfter := strings.Join(sortValues, ",")

			// A repeated cursor would page through the same results
			// forever.
			if nextAfter == after || nextAfter == "" {
				break
			}

			after = nextAfter
		}

		if page == sources.MaxPages {
			result := sources.WarningResult(s.Name(), fmt.Sprintf("pagination capped at %d pages", sources.MaxPages))

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
