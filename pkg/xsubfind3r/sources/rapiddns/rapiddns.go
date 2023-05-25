package rapiddns

import (
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

		res, err = httpclient.SimpleGet(fmt.Sprintf("https://rapiddns.io/subdomain/%s?full=1", config.Domain))
		if err != nil {
			return
		}

		for _, subdomain := range config.SubdomainsRegex.FindAllString(string(res.Body()), -1) {
			subdomains <- sources.Subdomain{Source: source.Name(), Value: subdomain}
		}
	}()

	return
}

func (source *Source) Name() string {
	return "rapiddns"
}
