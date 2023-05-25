package certspotterv0

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/valyala/fasthttp"
)

type response struct {
	ID       int      `json:"id"`
	DNSNames []string `json:"dns_names"`
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

		res, err = httpclient.SimpleGet(fmt.Sprintf("https://certspotter.com/api/v0/certs?domain=%s", config.Domain))
		if err != nil {
			return
		}

		var results []response

		if err = json.Unmarshal(res.Body(), &results); err != nil {
			return
		}

		for _, i := range results {
			for _, j := range i.DNSNames {
				subdomains <- sources.Subdomain{Source: source.Name(), Value: j}
			}
		}
	}()

	return
}

func (source *Source) Name() string {
	return "certspotterv0"
}
