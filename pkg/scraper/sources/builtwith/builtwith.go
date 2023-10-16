package builtwith

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hueristiq/xsubfind3r/pkg/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/scraper/sources"
)

type getDomainInfoResponse struct {
	Results []struct {
		Result struct {
			Paths []struct {
				Domain    string `json:"Domain"`
				URL       string `json:"Url"`
				SubDomain string `json:"SubDomain"`
			} `json:"Paths"`
		} `json:"Result"`
	} `json:"Results"`
}

type Source struct{}

func (source *Source) Run(config *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		var err error

		var key string

		key, err = sources.PickRandom(config.Keys.BuiltWith)
		if key == "" || err != nil {
			result := sources.Result{
				Type:   sources.Error,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		getDomainInfoReqURL := fmt.Sprintf("https://api.builtwith.com/v21/api.json?KEY=%s&HIDETEXT=yes&HIDEDL=yes&NOLIVE=yes&NOMETA=yes&NOPII=yes&NOATTR=yes&LOOKUP=%s", key, domain)

		var getDomainInfoRes *http.Response

		getDomainInfoRes, err = httpclient.SimpleGet(getDomainInfoReqURL)
		if err != nil {
			result := sources.Result{
				Type:   sources.Error,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			getDomainInfoRes.Body.Close()

			return
		}

		var getDomainInfoResData getDomainInfoResponse

		if err = json.NewDecoder(getDomainInfoRes.Body).Decode(&getDomainInfoResData); err != nil {
			result := sources.Result{
				Type:   sources.Error,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			getDomainInfoRes.Body.Close()

			return
		}

		getDomainInfoRes.Body.Close()

		for index := range getDomainInfoResData.Results {
			item := getDomainInfoResData.Results[index]

			for index := range item.Result.Paths {
				path := item.Result.Paths[index]

				value := path.Domain

				if path.SubDomain != "" {
					value = path.SubDomain + "." + value
				}

				result := sources.Result{
					Type:   sources.Subdomain,
					Source: source.Name(),
					Value:  value,
				}

				results <- result
			}
		}
	}()

	return results
}

func (source *Source) Name() string {
	return "builtwith"
}
