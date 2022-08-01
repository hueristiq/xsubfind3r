package certspotterv0

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/subfind3r/pkg/sources"
)

type response struct {
	ID       int      `json:"id"`
	DNSNames []string `json:"dns_names"`
}

type Source struct{}

func (source *Source) Run(domain string, session *sources.Session) chan sources.Subdomain {
	subdomains := make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		res, err := session.SimpleGet(fmt.Sprintf("https://certspotter.com/api/v0/certs?domain=%s", domain))
		if err != nil {
			return
		}

		var results []response

		if err := json.Unmarshal(res.Body(), &results); err != nil {
			return
		}

		for _, i := range results {
			for _, j := range i.DNSNames {
				subdomains <- sources.Subdomain{Source: source.Name(), Value: j}
			}
		}
	}()

	return subdomains
}

func (source *Source) Name() string {
	return "certspotterv0"
}
