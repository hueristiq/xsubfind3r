package sublist3r

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

		res, err = httpclient.SimpleGet(fmt.Sprintf("https://api.sublist3r.com/search.php?domain=%s", config.Domain))
		if err != nil {
			return
		}

		body := res.Body()

		var results []string

		if err = json.Unmarshal(body, &results); err != nil {
			return
		}

		for _, i := range results {
			subdomains <- sources.Subdomain{Source: source.Name(), Value: i}
		}
	}()

	return
}

func (source *Source) Name() string {
	return "sublist3r"
}
