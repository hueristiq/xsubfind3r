package alienvault

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/subfind3r/pkg/sources"
)

type response struct {
	Detail     string `json:"detail"`
	Error      string `json:"error"`
	PassiveDNS []struct {
		Hostname string `json:"hostname"`
	} `json:"passive_dns"`
}

type Source struct{}

func (source *Source) Run(domain string, session *sources.Session) chan sources.Subdomain {
	subdomains := make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		res, _ := session.SimpleGet(fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/domain/%s/passive_dns", domain))

		var results response

		if err := json.Unmarshal(res.Body(), &results); err != nil {
			return
		}

		if results.Error != "" {
			return
		}

		for _, j := range results.PassiveDNS {
			subdomains <- sources.Subdomain{Source: source.Name(), Value: j.Hostname}
		}
	}()

	return subdomains
}

func (source *Source) Name() string {
	return "alienvault"
}
