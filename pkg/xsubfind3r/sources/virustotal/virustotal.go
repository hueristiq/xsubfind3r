package virustotal

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	hqgohttp "go.source.hueristiq.com/http"
)

type getSubdomainsResponse struct {
	Data []struct {
		ID    string `json:"id"`
		Type  string `json:"type"`
		Links struct {
			Self string `json:"self"`
		} `json:"links"`
	} `json:"data"`
	Meta struct {
		Cursor string `json:"cursor"`
	} `json:"meta"`
}

type Source struct{}

func (source *Source) Run(cfg *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		key, err := cfg.Keys.VirusTotal.PickRandom()
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var cursor string

		for {
			getSubdomainsReqURL := fmt.Sprintf("https://www.virustotal.com/api/v3/domains/%s/subdomains", domain)
			getSubdomainsReqCFG := &hqgohttp.RequestConfiguration{
				Headers: map[string]string{
					"x-apikey": key,
				},
				Params: map[string]string{
					"limit": "1000",
				},
			}

			if cursor != "" {
				getSubdomainsReqCFG.Params["cursor"] = cursor
			}

			getSubdomainsRes, err := hqgohttp.Get(getSubdomainsReqURL, getSubdomainsReqCFG)
			if err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: source.Name(),
					Error:  err,
				}

				results <- result

				break
			}

			var getSubdomainsResData getSubdomainsResponse

			if err = json.NewDecoder(getSubdomainsRes.Body).Decode(&getSubdomainsResData); err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: source.Name(),
					Error:  err,
				}

				results <- result

				getSubdomainsRes.Body.Close()

				break
			}

			getSubdomainsRes.Body.Close()

			for _, record := range getSubdomainsResData.Data {
				subdomain := record.ID

				result := sources.Result{
					Type:   sources.ResultSubdomain,
					Source: source.Name(),
					Value:  subdomain,
				}

				results <- result
			}

			cursor = getSubdomainsResData.Meta.Cursor

			if cursor == "" {
				break
			}
		}
	}()

	return results
}

func (source *Source) Name() string {
	return sources.VIRUSTOTAL
}
