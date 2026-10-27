package netlas

import (
	"context"
	"encoding/json"
	"fmt"

	hqgohttpmethod "github.com/hueristiq/hq-lib-http-go/method"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

const communityDownloadCap = 200

type getCountResponse struct {
	Count int `json:"count"`
}

type downloadRequest struct {
	Query      string   `json:"q"`
	Fields     []string `json:"fields"`
	SourceType string   `json:"source_type"`
	Size       int      `json:"size"`
}

type downloadResponse []struct {
	Data struct {
		Domain string `json:"domain"`
	} `json:"data"`
}

type Source struct {
	keys sources.Keys
}

func (s *Source) Name() (name string) {
	name = sources.Netlas

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

		query := fmt.Sprintf("domain:*.%s AND NOT domain:%s", domain, domain)

		getCountReqURL := "https://app.netlas.io/api/domains_count/"

		getCountRes, err := cfg.Client.Request().
			URL(getCountReqURL).
			Param("q", query).
			SetHeader("Accept", "application/json").
			SetHeader("X-API-Key", key).
			Context(ctx).Do()
		if err != nil {
			result := sources.ErrorResult(s.Name(), fmt.Errorf("request failed: %w", err))

			results <- result

			return
		}

		if err = sources.CheckResponse(getCountRes.Std(), key); err != nil {
			result := sources.ErrorResult(s.Name(), err)

			results <- result

			_ = getCountRes.Body.Close()

			return
		}

		var getCountResData getCountResponse

		if err = json.NewDecoder(getCountRes.Body).Decode(&getCountResData); err != nil {
			result := sources.ErrorResult(s.Name(), fmt.Errorf("failed to parse JSON response: %w", err))

			results <- result

			_ = getCountRes.Body.Close()

			return
		}

		_ = getCountRes.Body.Close()

		size := getCountResData.Count

		if size > communityDownloadCap {
			size = communityDownloadCap
		}

		downloadReqBody, err := json.Marshal(downloadRequest{
			Query:      query,
			Fields:     []string{"*"},
			SourceType: "include",
			Size:       size,
		})
		if err != nil {
			result := sources.ErrorResult(s.Name(), fmt.Errorf("failed to marshal download request: %w", err))

			results <- result

			return
		}

		getSubdomainsReqURL := "https://app.netlas.io/api/domains/download/"

		getSubdomainsRes, err := cfg.Client.Request().
			Method(hqgohttpmethod.POST).
			URL(getSubdomainsReqURL).
			SetHeader("Accept", "application/json").
			SetHeader("X-API-Key", key).
			SetHeader("Content-Type", "application/json").
			Body(downloadReqBody).
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

		var getSubdomainsResData downloadResponse

		if err = json.NewDecoder(getSubdomainsRes.Body).Decode(&getSubdomainsResData); err != nil {
			result := sources.ErrorResult(s.Name(), fmt.Errorf("failed to parse JSON response: %w", err))

			results <- result

			_ = getSubdomainsRes.Body.Close()

			return
		}

		_ = getSubdomainsRes.Body.Close()

		for _, item := range getSubdomainsResData {
			subdomain, valid := cfg.Validator(item.Data.Domain)
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
