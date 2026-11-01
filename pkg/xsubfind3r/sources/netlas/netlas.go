package netlas

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

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

		getCountRes, err := func() (*http.Response, error) {
			req, err := http.NewRequestWithContext(ctx, http.MethodGet, getCountReqURL, http.NoBody)
			if err != nil {
				return nil, fmt.Errorf("build request: %w", err)
			}

			req.Header.Set("Accept", "application/json")
			req.Header.Set("X-Api-Key", key)
			queryValues := req.URL.Query()
			queryValues.Add("q", query)
			req.URL.RawQuery = queryValues.Encode()

			return cfg.Client.Do(req)
		}()
		if err != nil {
			result := sources.ErrorResult(s.Name(), fmt.Errorf("request failed: %w", err))

			results <- result

			return
		}

		if err = sources.CheckResponse(getCountRes, key); err != nil {
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

		getSubdomainsRes, err := func() (*http.Response, error) {
			var body io.Reader

			switch bodyValue := any(downloadReqBody).(type) {
			case nil:
			case string:
				body = strings.NewReader(bodyValue)
			case []byte:
				body = bytes.NewReader(bodyValue)
			case func() (io.Reader, error):
				var err error

				body, err = bodyValue()
				if err != nil {
					return nil, fmt.Errorf("build request: %w", err)
				}
			case io.Reader:
				body = bodyValue
			default:
				return nil, fmt.Errorf("unsupported request body type %T", bodyValue)
			}

			req, err := http.NewRequestWithContext(ctx, http.MethodPost, getSubdomainsReqURL, body)
			if err != nil {
				return nil, fmt.Errorf("build request: %w", err)
			}

			req.Header.Set("Accept", "application/json")
			req.Header.Set("X-Api-Key", key)
			req.Header.Set("Content-Type", "application/json")

			return cfg.Client.Do(req)
		}()
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
