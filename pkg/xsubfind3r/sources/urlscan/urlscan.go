package urlscan

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hueristiq/xsubfind3r/pkg/httpclient"
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

type Source struct{}

func (source *Source) Run(config *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		var err error

		var key string

		key, err = sources.PickRandom(config.Keys.URLScan)
		if err != nil {
			result := sources.Result{
				Type:   sources.Error,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		searchReqHeaders := map[string]string{
			"Content-Type": "application/json",
		}

		if key != "" {
			searchReqHeaders["API-Key"] = key
		}

		var searchAfter []interface{}

		for {
			after := ""

			if searchAfter != nil {
				searchAfterJSON, _ := json.Marshal(searchAfter)
				after = "&search_after=" + string(searchAfterJSON)
			}

			searchReqURL := fmt.Sprintf("https://urlscan.io/api/v1/search/?q=domain:%s&size=100", domain) + after

			var searchRes *http.Response

			searchRes, err = httpclient.Get(searchReqURL, "", searchReqHeaders)
			if err != nil {
				result := sources.Result{
					Type:   sources.Error,
					Source: source.Name(),
					Error:  err,
				}

				results <- result

				return
			}

			var searchResData searchResponse

			err = json.NewDecoder(searchRes.Body).Decode(&searchResData)
			if err != nil {
				result := sources.Result{
					Type:   sources.Error,
					Source: source.Name(),
					Error:  err,
				}

				results <- result

				searchRes.Body.Close()

				return
			}

			searchRes.Body.Close()

			if searchResData.Status == 429 {
				break
			}

			for _, result := range searchResData.Results {
				if !strings.HasSuffix(result.Page.Domain, "."+domain) {
					continue
				}

				result := sources.Result{
					Type:   sources.Subdomain,
					Source: source.Name(),
					Value:  result.Page.Domain,
				}

				results <- result
			}

			if !searchResData.HasMore {
				break
			}

			lastResult := searchResData.Results[len(searchResData.Results)-1]
			searchAfter = lastResult.Sort
		}
	}()

	return results
}

func (source *Source) Name() string {
	return "urlscan"
}
