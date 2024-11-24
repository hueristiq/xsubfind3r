package virustotal

import (
	"encoding/json"
	"fmt"

	hqgohttp "github.com/hueristiq/hq-go-http"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type getSubdomainsResponse struct {
	Data []struct {
		ID string `json:"id"`
	} `json:"data"`
	Meta struct {
		Cursor string `json:"cursor"`
	} `json:"meta"`
}

type Source struct{}

func (source *Source) Run(config *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		key, err := config.Keys.VirusTotal.PickRandom()
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
			getSubdomainsReqURL := fmt.Sprintf("https://www.virustotal.com/api/v3/domains/%s/subdomains?limit=1000", domain)

			if cursor != "" {
				getSubdomainsReqURL = fmt.Sprintf("%s&cursor=%s", getSubdomainsReqURL, cursor)
			}

			getSubdomainsRes, err := hqgohttp.GET(getSubdomainsReqURL).AddHeader("x-apikey", key).Send()
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
