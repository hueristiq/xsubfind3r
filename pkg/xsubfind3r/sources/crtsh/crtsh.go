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

func (source *Source) Run(config *sources.Configuration) (subdomains chan sources.Subdomain) {
	subdomains = make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		var (
			err error
			res *fasthttp.Response
		)

		res, err = httpclient.SimpleGet(fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", config.Domain))
		if err != nil {
			return
		}

		var results []response

		if err := json.Unmarshal(res.Body(), &results); err != nil {
			return
		}

		for _, i := range results {
			x := strings.Split(i.NameValue, "\n")

			for _, j := range x {
				subdomains <- sources.Subdomain{Source: source.Name(), Value: j}
			}
		}
	}()

	return
}

func (source *Source) Name() string {
	return "crtsh"
}
