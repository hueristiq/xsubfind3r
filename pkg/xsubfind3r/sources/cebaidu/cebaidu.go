package cebaidu

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/valyala/fasthttp"
)

type response struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    []struct {
		Domain string `json:"domain"`
	} `json:"data"`
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

		res, err = httpclient.SimpleGet(fmt.Sprintf("https://ce.baidu.com/index/getRelatedSites?site_address=%s", config.Domain))
		if err != nil {
			return
		}

		var results response

		if err = json.Unmarshal(res.Body(), &results); err != nil {
			return
		}

		for _, i := range results.Data {
			subdomains <- sources.Subdomain{Source: source.Name(), Value: i.Domain}
		}
	}()

	return
}

func (source *Source) Name() string {
	return "cebaidu"
}
