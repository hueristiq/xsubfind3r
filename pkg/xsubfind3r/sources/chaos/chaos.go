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

func (source *Source) Run(config *sources.Configuration, domain string) (subdomains chan sources.Subdomain) {
	subdomains = make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		var (
			key string
			err error
			res *fasthttp.Response
		)

		key, err = sources.PickRandom(config.Keys.Chaos)
		if key == "" || err != nil {
			return
		}

		reqURL := fmt.Sprintf("https://dns.projectdiscovery.io/dns/%s/subdomains", domain)
		headers := map[string]string{"Authorization": key}

		res, err = httpclient.Request(fasthttp.MethodGet, reqURL, "", headers, nil)
		if err != nil {
			return
		}

		var results response

		if err = json.Unmarshal(res.Body(), &results); err != nil {
			return
		}

		for _, record := range results.Subdomains {
			subdomain := fmt.Sprintf("%s.%s", record, results.Domain)

			subdomains <- sources.Subdomain{Source: source.Name(), Value: subdomain}
		}
	}()

	return
}

func (source *Source) Name() string {
	return "chaos"
}
