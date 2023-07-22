package crtsh

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/valyala/fasthttp"
)

type response struct {
	ID        int    `json:"id"`
	NameValue string `json:"name_value"`
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

		reqURL := fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", domain)

		res, err = httpclient.SimpleGet(reqURL)
		if err != nil {
			return
		}

		var results []response

		if err := json.Unmarshal(res.Body(), &results); err != nil {
			return
		}

		for _, record := range results {
			for _, subdomain := range strings.Split(record.NameValue, "\n") {
				subdomains <- sources.Subdomain{Source: source.Name(), Value: subdomain}
			}
		}
	}()

	return
}

func (source *Source) Name() string {
	return "crtsh"
}
