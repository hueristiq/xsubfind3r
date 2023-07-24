package shodan

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/valyala/fasthttp"
)

// type response struct {
// 	ID        int    `json:"id"`
// 	NameValue string `json:"name_value"`
// }

type response struct {
	Domain     string   `json:"domain"`
	Subdomains []string `json:"subdomains"`
	Result     int      `json:"result"`
	Error      string   `json:"error"`
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

		key, err = sources.PickRandom(config.Keys.Shodan)
		if key == "" || err != nil {
			return
		}

		reqURL := fmt.Sprintf("https://api.shodan.io/dns/domain/%s?key=%s", domain, key)

		res, err = httpclient.SimpleGet(reqURL)
		if err != nil {
			return
		}

		var results response

		if err := json.Unmarshal(res.Body(), &results); err != nil {
			return
		}

		for _, subdomain := range results.Subdomains {
			subdomains <- sources.Subdomain{Source: source.Name(), Value: fmt.Sprintf("%s.%s", subdomain, domain)}
		}
	}()

	return
}

func (source *Source) Name() string {
	return "shodan"
}
