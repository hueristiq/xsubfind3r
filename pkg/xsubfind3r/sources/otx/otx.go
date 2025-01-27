package otx

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	hqgohttp "go.source.hueristiq.com/http"
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

		getPassiveDNSReqURL := fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/domain/%s/passive_dns", domain)

		getPassiveDNSRes, err := hqgohttp.GET(getPassiveDNSReqURL).Send()
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var getPassiveDNSResData getPassiveDNSResponse

		if err = json.NewDecoder(getPassiveDNSRes.Body).Decode(&getPassiveDNSResData); err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
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
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  fmt.Errorf("%w: %s, %s", errStatic, getPassiveDNSResData.Detail, getPassiveDNSResData.Error),
			}

			results <- result

			return
		}

		for _, record := range getPassiveDNSResData.PassiveDNS {
			subdomain := record.Hostname

			if subdomain != domain && !strings.HasSuffix(subdomain, "."+domain) {
				continue
			}

			result := sources.Result{
				Type:   sources.ResultSubdomain,
				Source: source.Name(),
				Value:  subdomain,
			}

			results <- result
		}
	}()

	return results
}

func (source *Source) Name() string {
	return sources.OPENTHREATEXCHANGE
}

var errStatic = errors.New("something went wrong")
