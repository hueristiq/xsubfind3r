package fullhunt

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/valyala/fasthttp"
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

		var err error

		var key string

		key, err = sources.PickRandom(config.Keys.Fullhunt)
		if key == "" || err != nil {
			result := sources.Result{
				Type:   sources.Error,
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

		var getSubdomainsRes *fasthttp.Response

		getSubdomainsRes, err = httpclient.Get(getSubdomainsReqURL, "", getSubdomainsReqHeaders)
		if err != nil {
			result := sources.Result{
				Type:   sources.Error,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var getSubdomainsResData getSubdomainsResponse

		err = json.Unmarshal(getSubdomainsRes.Body(), &getSubdomainsResData)
		if err != nil {
			result := sources.Result{
				Type:   sources.Error,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		for _, subdomain := range getSubdomainsResData.Hosts {
			result := sources.Result{
				Type:   sources.Subdomain,
				Source: source.Name(),
				Value:  subdomain,
			}

			results <- result
		}
	}()

	return results
}

func (source *Source) Name() string {
	return "fullhunt"
}
