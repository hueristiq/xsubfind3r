package chaos

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/valyala/fasthttp"
)

type getSubdomainsResponse struct {
	Domain     string   `json:"domain"`
	Subdomains []string `json:"subdomains"`
	Count      int      `json:"count"`
}

type Source struct{}

func (source *Source) Run(config *sources.Configuration, domain string) (subdomainsChannel chan sources.Subdomain) {
	subdomainsChannel = make(chan sources.Subdomain)

	go func() {
		defer close(subdomainsChannel)

		var err error

		var key string

		key, err = sources.PickRandom(config.Keys.Chaos)
		if key == "" || err != nil {
			return
		}

		getSubdomainsReqHeaders := map[string]string{"Authorization": key}

		getSubdomainsReqURL := fmt.Sprintf("https://dns.projectdiscovery.io/dns/%s/subdomains", domain)

		var getSubdomainsRes *fasthttp.Response

		getSubdomainsRes, err = httpclient.Get(getSubdomainsReqURL, "", getSubdomainsReqHeaders)
		if err != nil {
			return
		}

		var getSubdomainsResData getSubdomainsResponse

		if err = json.Unmarshal(getSubdomainsRes.Body(), &getSubdomainsResData); err != nil {
			return
		}

		for _, record := range getSubdomainsResData.Subdomains {
			subdomain := fmt.Sprintf("%s.%s", record, getSubdomainsResData.Domain)

			subdomainsChannel <- sources.Subdomain{Source: source.Name(), Value: subdomain}
		}
	}()

	return
}

func (source *Source) Name() string {
	return "chaos"
}
