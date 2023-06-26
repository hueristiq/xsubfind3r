package alienvault

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/valyala/fasthttp"
)

type response struct {
	Detail     string `json:"detail"`
	Error      string `json:"error"`
	PassiveDNS []struct {
		Hostname string `json:"hostname"`
	} `json:"passive_dns"`
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

		reqURL := fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/domain/%s/passive_dns", config.Domain)

		res, err = httpclient.SimpleGet(reqURL)
		if err != nil {
			return
		}

		var results response

		if err = json.Unmarshal(res.Body(), &results); err != nil {
			return
		}

		if results.Error != "" {
			return
		}

		for _, record := range results.PassiveDNS {
			subdomains <- sources.Subdomain{Source: source.Name(), Value: record.Hostname}
		}
	}()

	return
}

func (source *Source) Name() string {
	return "alienvault"
}
