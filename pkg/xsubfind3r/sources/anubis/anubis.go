package anubis

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/valyala/fasthttp"
)

type Source struct{}

func (source *Source) Run(config *sources.Configuration) (subdomains chan sources.Subdomain) {
	subdomains = make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		var (
			err error
			res *fasthttp.Response
		)

		reqURL := fmt.Sprintf("https://jldc.me/anubis/subdomains/%s", config.Domain)

		res, err = httpclient.SimpleGet(reqURL)
		if err != nil {
			return
		}

		var results []string

		if err = json.Unmarshal(res.Body(), &results); err != nil {
			return
		}

		for _, subdomain := range results {
			subdomains <- sources.Subdomain{Source: source.Name(), Value: subdomain}
		}
	}()

	return
}

func (source *Source) Name() string {
	return "anubis"
}
