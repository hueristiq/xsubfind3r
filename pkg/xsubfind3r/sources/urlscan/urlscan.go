package urlscan

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/valyala/fasthttp"
)

type response struct {
	Results []struct {
		Page struct {
			Domain string `json:"domain"`
		} `json:"page"`
	} `json:"results"`
}

type Source struct{}

func (source *Source) Run(config *sources.Configuration) (subdomains chan sources.Subdomain) {
	subdomains = make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		var (
			err error
			res *fasthttp.Response
		)

		res, err = httpclient.SimpleGet(fmt.Sprintf("https://urlscan.io/api/v1/search/?q=domain:%s", config.Domain))
		if err != nil {
			return
		}

		body := res.Body()

		var results response

		if err = json.Unmarshal(body, &results); err != nil {
			return
		}

		for _, i := range results.Results {
			subdomains <- sources.Subdomain{Source: source.Name(), Value: i.Page.Domain}
		}
	}()

	return
}

func (source *Source) Name() string {
	return "urlscan"
}
