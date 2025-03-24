package intelx

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	hqgohttp "go.source.hueristiq.com/http"
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

type Source struct{}

func (source *Source) Run(cfg *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		key, err := cfg.Keys.Intelx.PickRandom()
		if key == "" || err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		parts := strings.Split(key, ":")
		if len(parts) != 2 {
			return
		}

		intelXHost := parts[0]
		intelXKey := parts[1]

		if intelXKey == "" || intelXHost == "" {
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
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		searchReqBodyReader := bytes.NewBuffer(searchReqBodyBytes)

		searchReqCFG := &hqgohttp.RequestConfiguration{
			Headers: map[string]string{
				"Content-Type": "application/json",
			},
		}

		var searchRes *http.Response

		searchRes, err = hqgohttp.Post(searchReqURL, searchReqBodyReader, searchReqCFG)
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var searchResData searchResponse

		if err = json.NewDecoder(searchRes.Body).Decode(&searchResData); err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			searchRes.Body.Close()

			return
		}

		searchRes.Body.Close()

		getResultsReqURL := fmt.Sprintf("https://%s/phonebook/search/result", intelXHost)
		getResultsReqCFG := &hqgohttp.RequestConfiguration{
			Params: map[string]string{
				"k":     intelXKey,
				"id":    searchResData.ID,
				"limit": "10000",
			},
		}
		status := 0

		for status == 0 || status == 3 {
			var getResultsRes *http.Response

			getResultsRes, err = hqgohttp.Get(getResultsReqURL, getResultsReqCFG)
			if err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: source.Name(),
					Error:  err,
				}

				results <- result

				return
			}

			var getResultsResData getResultsResponse

			if err = json.NewDecoder(getResultsRes.Body).Decode(&getResultsResData); err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: source.Name(),
					Error:  err,
				}

				results <- result

				getResultsRes.Body.Close()

				return
			}

			getResultsRes.Body.Close()

			status = getResultsResData.Status

			for _, record := range getResultsResData.Selectors {
				result := sources.Result{
					Type:   sources.ResultSubdomain,
					Source: source.Name(),
					Value:  record.Selectvalue,
				}

				results <- result
			}
		}
	}()

	return results
}

func (source *Source) Name() string {
	return sources.INTELLIGENCEX
}
