package shodan

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/valyala/fasthttp"
)

type getDNSResponse struct {
	Domain     string   `json:"domain"`
	Subdomains []string `json:"subdomains"`
	Result     int      `json:"result"`
	Error      string   `json:"error"`
}

type Source struct{}

func (source *Source) Run(config *sources.Configuration, domain string) (subdomainsChannel chan sources.Subdomain) {
	subdomainsChannel = make(chan sources.Subdomain)

	go func() {
		defer close(subdomainsChannel)

		var err error

		var key string

		key, err = sources.PickRandom(config.Keys.Shodan)
		if key == "" || err != nil {
			return
		}

		getDNSReqURL := fmt.Sprintf("https://api.shodan.io/dns/domain/%s?key=%s", domain, key)

		var getDNSRes *fasthttp.Response

		getDNSRes, err = httpclient.SimpleGet(getDNSReqURL)
		if err != nil {
			return
		}

		var getDNSResData getDNSResponse

		if err := json.Unmarshal(getDNSRes.Body(), &getDNSResData); err != nil {
			return
		}

		for _, subdomain := range getDNSResData.Subdomains {
			subdomainsChannel <- sources.Subdomain{Source: source.Name(), Value: fmt.Sprintf("%s.%s", subdomain, domain)}
		}
	}()

	return
}

func (source *Source) Name() string {
	return "shodan"
}
