package intelx

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	hqgohttpheader "github.com/hueristiq/hq-lib-http-go/header"
	hqgohttpmethod "github.com/hueristiq/hq-lib-http-go/method"
	hqgohttpmime "github.com/hueristiq/hq-lib-http-go/mime"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

const (
	// pollInterval is the wait between result polls while a search is
	// still running on the API side.
	pollInterval = 2 * time.Second
	// maxPollAttempts caps the result polls before giving up.
	maxPollAttempts = 30
)

type searchRequestBody struct {
	Term       string        `json:"term"`
	MaxResults int           `json:"maxresults"`
	Media      int           `json:"media"`
	Target     int           `json:"target"`
	Timeout    time.Duration `json:"timeout"`
}

type searchResponse struct {
	ID                string `json:"id"`
	SelfSelectWarning bool   `json:"selfselectwarning"`
	Status            int    `json:"status"`
	AltTerm           string `json:"altterm"`
	AltTermH          string `json:"alttermh"`
}

type getResultsResponse struct {
	Selectors []struct {
		Selectvalue string `json:"selectorvalue"`
	} `json:"selectors"`
	Status int `json:"status"`
}

type Source struct {
	keys sources.Keys
}

func (s *Source) Name() (name string) {
	name = sources.INTELLIGENCEX

	return
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

		parts := strings.Split(key, ":")
		if len(parts) != 2 {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  errors.New("malformed intelx key: expected the \"host:key\" format"),
			}

			results <- result

			return
		}

		intelXHost := parts[0]
		intelXKey := parts[1]

		if intelXKey == "" || intelXHost == "" {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  errors.New("malformed intelx key: expected the \"host:key\" format"),
			}

			results <- result

			return
		}

		searchReqURL := fmt.Sprintf("https://%s/phonebook/search?k=%s", intelXHost, intelXKey)
		searchReqBody := searchRequestBody{
			Term:       domain,
			MaxResults: 100000,
			Media:      0,
			Target:     1, // 1 = Domains | 2 = Emails | 3 = URLs
			Timeout:    20,
		}

		var searchReqBodyBytes []byte

		searchReqBodyBytes, err = json.Marshal(searchReqBody)
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		searchReqBodyReader := bytes.NewBuffer(searchReqBodyBytes)

		searchRes, err := cfg.Client.NewRequest().
			Method(hqgohttpmethod.POST).
			URL(searchReqURL).
			SetHeader(hqgohttpheader.ContentType.String(), hqgohttpmime.JSON.String()).
			Body(searchReqBodyReader).
			Do(ctx)
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  fmt.Errorf("request failed: %w", sources.SanitizeError(err, intelXKey)),
			}

			results <- result

			return
		}

		var searchResData searchResponse

		if err = json.NewDecoder(searchRes.Body).Decode(&searchResData); err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  fmt.Errorf("failed to parse JSON response: %w", err),
			}

			results <- result

			sources.CloseBody(searchRes.Body)

			return
		}

		sources.CloseBody(searchRes.Body)

		getResultsReqURL := fmt.Sprintf("https://%s/phonebook/search/result", intelXHost)
		status := 0
		polls := 0

		for status == 0 || status == 3 {
			polls++

			if polls > maxPollAttempts {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: s.Name(),
					Error:  fmt.Errorf("search timed out after %d polls", maxPollAttempts),
				}

				results <- result

				return
			}

			getResultsRes, err := cfg.Client.NewRequest().
				URL(getResultsReqURL).
				Param("k", intelXKey).
				Param("id", searchResData.ID).
				Param("limit", "10000").
				Do(ctx)
			if err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: s.Name(),
					Error:  fmt.Errorf("request failed: %w", sources.SanitizeError(err, intelXKey)),
				}

				results <- result

				return
			}

			var getResultsResData getResultsResponse

			if err = json.NewDecoder(getResultsRes.Body).Decode(&getResultsResData); err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: s.Name(),
					Error:  fmt.Errorf("failed to parse JSON response: %w", err),
				}

				results <- result

				sources.CloseBody(getResultsRes.Body)

				return
			}

			sources.CloseBody(getResultsRes.Body)

			status = getResultsResData.Status

			for _, record := range getResultsResData.Selectors {
				result := sources.Result{
					Type:   sources.ResultSubdomain,
					Source: s.Name(),
					Value:  record.Selectvalue,
				}

				results <- result
			}

			if status == 0 || status == 3 {
				select {
				case <-ctx.Done():
					return
				case <-time.After(pollInterval):
				}
			}
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
