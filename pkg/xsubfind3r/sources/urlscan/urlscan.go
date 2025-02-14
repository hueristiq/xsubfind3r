package urlscan

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/spf13/cast"
	hqgohttp "go.source.hueristiq.com/http"
	"go.source.hueristiq.com/http/method"
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

type Source struct{}

func (source *Source) Run(cfg *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		key, err := cfg.Keys.URLScan.PickRandom()
		if err != nil && !errors.Is(err, sources.ErrNoKeys) {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var after string

		for {
			searchReqURL := fmt.Sprintf("https://urlscan.io/api/v1/search/?q=domain:%s&size=10000", domain)

			if after != "" {
				searchReqURL += "&search_after=" + after
			}

			var searchRes *http.Response

			searchRes, err = hqgohttp.Request().Method(method.GET.String()).URL(searchReqURL).AddHeader("Content-Type", "application/json").AddHeader("API-Key", key).Send()
			if err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: source.Name(),
					Error:  err,
				}

				results <- result

				break
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

				break
			}

			searchRes.Body.Close()

			if searchResData.Status == 429 {
				break
			}

			for _, record := range searchResData.Results {
				subdomain := record.Page.Domain

				if subdomain != domain && !strings.HasSuffix(subdomain, "."+domain) {
					continue
				}

				result := sources.Result{
					Type:   sources.ResultSubdomain,
					Source: source.Name(),
					Value:  subdomain,
				}

				results <- result
			}

			if !searchResData.HasMore {
				break
			}

			if len(searchResData.Results) < 1 {
				break
			}

			lastResult := searchResData.Results[len(searchResData.Results)-1]

			if lastResult.Sort != nil {
				var temp []string

				for index := range lastResult.Sort {
					temp = append(temp, cast.ToString(lastResult.Sort[index]))
				}

				after = strings.Join(temp, ",")
			}
		}
	}()

	return results
}

func (source *Source) Name() string {
	return sources.URLSCAN
}
