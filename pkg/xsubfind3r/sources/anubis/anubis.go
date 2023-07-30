package anubis

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/valyala/fasthttp"
)

type Source struct{}

func (source *Source) Run(_ *sources.Configuration, domain string) (subdomainsChannel chan sources.Subdomain) {
	subdomainsChannel = make(chan sources.Subdomain)

	go func() {
		defer close(subdomainsChannel)

		var err error

		getSubdomainsReqURL := fmt.Sprintf("https://jldc.me/anubis/subdomains/%s", domain)

		var getSubdomainsRes *fasthttp.Response

		getSubdomainsRes, err = httpclient.SimpleGet(getSubdomainsReqURL)
		if err != nil {
			return
		}

		var getSubdomainsResData []string

		if err = json.Unmarshal(getSubdomainsRes.Body(), &getSubdomainsResData); err != nil {
			return
		}

		for _, subdomain := range getSubdomainsResData {
			subdomainsChannel <- sources.Subdomain{Source: source.Name(), Value: subdomain}
		}
	}()

	return
}

func (source *Source) Name() string {
	return "anubis"
}
