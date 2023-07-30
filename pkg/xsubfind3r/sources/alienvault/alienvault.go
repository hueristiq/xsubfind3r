package alienvault

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/valyala/fasthttp"
)

type getPassiveDNSResponse struct {
	Detail     string `json:"detail"`
	Error      string `json:"error"`
	PassiveDNS []struct {
		Hostname string `json:"hostname"`
	} `json:"passive_dns"`
}

type Source struct{}

func (source *Source) Run(_ *sources.Configuration, domain string) (subdomainsChannel chan sources.Subdomain) {
	subdomainsChannel = make(chan sources.Subdomain)

	go func() {
		defer close(subdomainsChannel)

		var err error

		getPassiveDNSReqURL := fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/domain/%s/passive_dns", domain)

		var getPassiveDNSRes *fasthttp.Response

		getPassiveDNSRes, err = httpclient.SimpleGet(getPassiveDNSReqURL)
		if err != nil {
			return
		}

		var getPassiveDNSResData getPassiveDNSResponse

		if err = json.Unmarshal(getPassiveDNSRes.Body(), &getPassiveDNSResData); err != nil {
			return
		}

		if getPassiveDNSResData.Error != "" {
			return
		}

		for _, record := range getPassiveDNSResData.PassiveDNS {
			subdomainsChannel <- sources.Subdomain{Source: source.Name(), Value: record.Hostname}
		}
	}()

	return
}

func (source *Source) Name() string {
	return "alienvault"
}
