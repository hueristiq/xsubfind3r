package archiveis

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

		res, err = httpclient.SimpleGet(fmt.Sprintf("https://archive.is/*.%s", config.Domain))
		if err != nil {
			return
		}

		src := string(res.Body())

		for _, subdomain := range config.SubdomainsRegex.FindAllString(src, -1) {
			subdomains <- sources.Subdomain{Source: source.Name(), Value: subdomain}
		}
	}()

	return
}

func (source *Source) Name() string {
	return "archiveis"
}
