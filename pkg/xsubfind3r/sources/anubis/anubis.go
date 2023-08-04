package anubis

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/valyala/fasthttp"
)

type Source struct{}

func (source *Source) Run(_ *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		var err error

		getSubdomainsReqURL := fmt.Sprintf("https://jldc.me/anubis/subdomains/%s", domain)

		var getSubdomainsRes *fasthttp.Response

		getSubdomainsRes, err = httpclient.SimpleGet(getSubdomainsReqURL)
		if err != nil {
			result := sources.Result{
				Type:   sources.Error,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var getSubdomainsResData []string

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

		for _, subdomain := range getSubdomainsResData {
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
	return "anubis"
}
