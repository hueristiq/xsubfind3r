package sonar

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/valyala/fasthttp"
)

// type response struct {
// 	Subdomains []string
// }

type Source struct{}

func (source *Source) Run(config *sources.Configuration) (subdomains chan sources.Subdomain) {
	subdomains = make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		var (
			err error
			res *fasthttp.Response
		)

		res, err = httpclient.SimpleGet(fmt.Sprintf("https://sonar.omnisint.io/subdomains/%s", config.Domain))
		if err != nil {
			return
		}

		var results []interface{}

		if err = json.Unmarshal(res.Body(), &results); err != nil {
			return
		}

		for _, i := range results {
			subdomains <- sources.Subdomain{Source: source.Name(), Value: i.(string)}
		}
	}()

	return
}

func (source *Source) Name() string {
	return "sonar"
}
