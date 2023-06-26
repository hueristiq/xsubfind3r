package bevigil

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/valyala/fasthttp"
)

type response struct {
	Domain     string   `json:"domain"`
	Subdomains []string `json:"subdomains"`
}

type Source struct{}

func (source *Source) Run(config *sources.Configuration) (subdomains chan sources.Subdomain) {
	subdomains = make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		var (
			key     string
			err     error
			res     *fasthttp.Response
			headers = map[string]string{}
		)

		key, err = sources.PickRandom(config.Keys.Bevigil)
		if key == "" || err != nil {
			return
		}

		if len(config.Keys.Bevigil) > 0 {
			headers["X-Access-Token"] = key
		}

		reqURL := fmt.Sprintf("https://osint.bevigil.com/api/%s/subdomains/", config.Domain)

		res, err = httpclient.Request(fasthttp.MethodGet, reqURL, "", headers, nil)
		if err != nil {
			return
		}

		body := res.Body()

		var results response

		if err = json.Unmarshal(body, &results); err != nil {
			return
		}

		for _, subdomain := range results.Subdomains {
			subdomains <- sources.Subdomain{Source: source.Name(), Value: subdomain}
		}
	}()

	return
}

func (source *Source) Name() string {
	return "bevigil"
}
