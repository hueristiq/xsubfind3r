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

func (source *Source) Run(_ *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		var err error

		getPassiveDNSReqURL := fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/domain/%s/passive_dns", domain)

		var getPassiveDNSRes *fasthttp.Response

		getPassiveDNSRes, err = httpclient.SimpleGet(getPassiveDNSReqURL)
		if err != nil {
			result := sources.Result{
				Type:   sources.Error,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var getPassiveDNSResData getPassiveDNSResponse

		err = json.Unmarshal(getPassiveDNSRes.Body(), &getPassiveDNSResData)
		if err != nil {
			result := sources.Result{
				Type:   sources.Error,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		if getPassiveDNSResData.Error != "" {
			return
		}

		for _, record := range getPassiveDNSResData.PassiveDNS {
			result := sources.Result{
				Type:   sources.Subdomain,
				Source: source.Name(),
				Value:  record.Hostname,
			}

			results <- result
		}
	}()

	return results
}

func (source *Source) Name() string {
	return "alienvault"
}
