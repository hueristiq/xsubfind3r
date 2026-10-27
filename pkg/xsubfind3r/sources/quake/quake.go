package quake

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	hqgohttpmethod "github.com/hueristiq/hq-lib-http-go/method"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type searchRequest struct {
	Query   string   `json:"query"`
	Include []string `json:"include"`
	Latest  bool     `json:"latest"`
	Size    int      `json:"size"`
	Start   int      `json:"start"`
}

type searchResponse struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Data    []struct {
		Service struct {
			HTTP struct {
				Host string `json:"host"`
			} `json:"http"`
		} `json:"service"`
	} `json:"data"`
	Meta struct {
		Pagination struct {
			Total int `json:"total"`
		} `json:"pagination"`
	} `json:"meta"`
}

type Source struct {
	keys sources.Keys
}

func (s *Source) Name() (name string) {
	name = sources.Quake

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

		pageSize := 500
		start := 0
		total := -1

		for {
			searchReqBody, err := json.Marshal(searchRequest{
				Query:   "domain: " + domain,
				Include: []string{"service.http.host"},
				Latest:  true,
				Size:    pageSize,
				Start:   start,
			})
			if err != nil {
				result := sources.ErrorResult(s.Name(), fmt.Errorf("failed to marshal search request: %w", err))

				results <- result

				return
			}

			getSubdomainsReqURL := "https://quake.360.net/api/v3/search/quake_service"

			getSubdomainsRes, err := cfg.Client.Request().
				Method(hqgohttpmethod.POST).
				URL(getSubdomainsReqURL).
				SetHeader("Content-Type", "application/json").
				SetHeader("X-QuakeToken", key).
				Body(searchReqBody).
				Context(ctx).Do()
			if err != nil {
				result := sources.ErrorResult(s.Name(), fmt.Errorf("request failed: %w", err))

				results <- result

				return
			}

			if err = sources.CheckResponse(getSubdomainsRes.Std(), key); err != nil {
				result := sources.ErrorResult(s.Name(), err)

				results <- result

				_ = getSubdomainsRes.Body.Close()

				return
			}

			var getSubdomainsResData searchResponse

			if err = json.NewDecoder(getSubdomainsRes.Body).Decode(&getSubdomainsResData); err != nil {
				result := sources.ErrorResult(s.Name(), fmt.Errorf("failed to parse JSON response: %w", err))

				results <- result

				_ = getSubdomainsRes.Body.Close()

				return
			}

			_ = getSubdomainsRes.Body.Close()

			if getSubdomainsResData.Code != 0 {
				result := sources.ErrorResult(s.Name(), fmt.Errorf("%s", getSubdomainsResData.Message))

				results <- result

				return
			}

			if total == -1 {
				total = getSubdomainsResData.Meta.Pagination.Total
			}

			for _, record := range getSubdomainsResData.Data {
				subdomain := record.Service.HTTP.Host

				// Quake returns this literal (Chinese for "no permission") in
				// the host field when the query is unauthorized.
				if strings.Contains(subdomain, "\u6682\u65e0\u6743\u9650") {
					continue
				}

				subdomain, valid := cfg.Validator(subdomain)
				if !valid {
					continue
				}

				result := sources.SubdomainResult(s.Name(), subdomain)

				results <- result
			}

			if len(getSubdomainsResData.Data) == 0 || start+pageSize >= total {
				break
			}

			start += pageSize
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
