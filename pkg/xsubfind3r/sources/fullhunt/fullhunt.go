package fullhunt

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hueristiq/xsubfind3r/pkg/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type getSubdomainsResponse struct {
	Hosts   []string `json:"hosts"`
	Message string   `json:"message"`
	Status  int      `json:"status"`
}

type Source struct{}

func (source *Source) Run(config *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		key, err := config.Keys.Fullhunt.PickRandom()
		if key == "" || err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		getSubdomainsReqHeaders := map[string]string{}

		if len(config.Keys.Fullhunt) > 0 {
			getSubdomainsReqHeaders["X-API-KEY"] = key
		}

		getSubdomainsReqURL := fmt.Sprintf("https://fullhunt.io/api/v1/domain/%s/subdomains", domain)

		var getSubdomainsRes *http.Response

		getSubdomainsRes, err = httpclient.Get(getSubdomainsReqURL, "", getSubdomainsReqHeaders)
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			httpclient.DiscardResponse(getSubdomainsRes)

			return
		}

		var getSubdomainsResData getSubdomainsResponse

		err = json.NewDecoder(getSubdomainsRes.Body).Decode(&getSubdomainsResData)
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			getSubdomainsRes.Body.Close()

			return
		}

		getSubdomainsRes.Body.Close()

		for _, subdomain := range getSubdomainsResData.Hosts {
			result := sources.Result{
				Type:   sources.ResultSubdomain,
				Source: source.Name(),
				Value:  subdomain,
			}

			results <- result
		}
	}()

	return results
}

func (source *Source) Name() string {
	return sources.FULLHUNT
}
