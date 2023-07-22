package urlscan

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/valyala/fasthttp"
)

type response struct {
	Results []struct {
		Page struct {
			Domain string `json:"domain"`
		} `json:"page"`
	} `json:"results"`
}

type Source struct{}

func (source *Source) Run(_ *sources.Configuration, domain string) (subdomains chan sources.Subdomain) {
	subdomains = make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		var (
			err error
			res *fasthttp.Response
		)

		reqURL := fmt.Sprintf("https://urlscan.io/api/v1/search/?q=domain:%s", domain)

		res, err = httpclient.SimpleGet(reqURL)
		if err != nil {
			return
		}

		body := res.Body()

		var results response

		if err = json.Unmarshal(body, &results); err != nil {
			return
		}

		for _, record := range results.Results {
			if !strings.HasSuffix(record.Page.Domain, "."+domain) {
				continue
			}

			subdomains <- sources.Subdomain{Source: source.Name(), Value: record.Page.Domain}
		}
	}()

	return
}

func (source *Source) Name() string {
	return "urlscan"
}
