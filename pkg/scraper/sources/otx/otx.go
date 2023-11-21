package otx

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hueristiq/xsubfind3r/pkg/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/scraper/sources"
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

		var getPassiveDNSRes *http.Response

		getPassiveDNSRes, err = httpclient.SimpleGet(getPassiveDNSReqURL)
		if err != nil {
			result := sources.Result{
				Type:   sources.Error,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			getPassiveDNSRes.Body.Close()

			return
		}

		var getPassiveDNSResData getPassiveDNSResponse

		if err = json.NewDecoder(getPassiveDNSRes.Body).Decode(&getPassiveDNSResData); err != nil {
			result := sources.Result{
				Type:   sources.Error,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			getPassiveDNSRes.Body.Close()

			return
		}

		getPassiveDNSRes.Body.Close()

		if getPassiveDNSResData.Error != "" {
			result := sources.Result{
				Type:   sources.Error,
				Source: source.Name(),
				Error:  fmt.Errorf("%s, %s", getPassiveDNSResData.Detail, getPassiveDNSResData.Error),
			}

			results <- result

			return
		}

		for index := range getPassiveDNSResData.PassiveDNS {
			subdomain := getPassiveDNSResData.PassiveDNS[index].Hostname

			if subdomain != domain && !strings.HasSuffix(subdomain, "."+domain) {
				continue
			}

			result := sources.Result{
				Type:   sources.Subdomain,
				Source: source.Name(),
				Value:  subdomain,
			}

			results <- result
		}
	}()

	return results
}

func (source *Source) Name() string {
	return "otx"
}
