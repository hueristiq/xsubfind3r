package chaos

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
	Count      int      `json:"count"`
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

		if config.Keys.Chaos == "" {
			return
		}

		res, err = httpclient.Request(
			fasthttp.MethodGet,
			fmt.Sprintf("https://dns.projectdiscovery.io/dns/%s/subdomains", config.Domain),
			"",
			map[string]string{"Authorization": config.Keys.Chaos},
			nil,
		)
		if err != nil {
			return
		}

		var results response

		if err = json.Unmarshal(res.Body(), &results); err != nil {
			return
		}

		for _, i := range results.Subdomains {
			subdomains <- sources.Subdomain{Source: source.Name(), Value: fmt.Sprintf("%s.%s", i, results.Domain)}
		}
	}()

	return
}

func (source *Source) Name() string {
	return "chaos"
}
