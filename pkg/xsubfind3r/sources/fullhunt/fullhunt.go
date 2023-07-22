package fullhunt

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/valyala/fasthttp"
)

type response struct {
	Hosts   []string `json:"hosts"`
	Message string   `json:"message"`
	Status  int      `json:"status"`
}

type Source struct{}

func (source *Source) Run(config *sources.Configuration, domain string) (subdomains chan sources.Subdomain) {
	subdomains = make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		var (
			key     string
			err     error
			res     *fasthttp.Response
			headers = map[string]string{}
		)

		key, err = sources.PickRandom(config.Keys.Fullhunt)
		if key == "" || err != nil {
			return
		}

		if len(config.Keys.Fullhunt) > 0 {
			headers["X-API-KEY"] = key
		}

		reqURL := fmt.Sprintf("https://fullhunt.io/api/v1/domain/%s/subdomains", domain)

		res, err = httpclient.Request(fasthttp.MethodGet, reqURL, "", headers, nil)
		if err != nil {
			return
		}

		body := res.Body()

		var results response

		if err = json.Unmarshal(body, &results); err != nil {
			return
		}

		for _, subdomain := range results.Hosts {
			subdomains <- sources.Subdomain{Source: source.Name(), Value: subdomain}
		}
	}()

	return
}

func (source *Source) Name() string {
	return "fullhunt"
}
