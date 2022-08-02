package threatcrowd

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/hqsubfind3r/pkg/sources"
)

type response struct {
	Subdomains []string `json:"subdomains"`
}

type Source struct{}

func (source *Source) Run(domain string, session *sources.Session) chan sources.Subdomain {
	subdomains := make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		res, err := session.SimpleGet(fmt.Sprintf("https://www.threatcrowd.org/searchApi/v2/domain/report/?domain=%s", domain))
		if err != nil {
			return
		}

		var results response

		if err := json.Unmarshal(res.Body(), &results); err != nil {
			return
		}

		for _, i := range results.Subdomains {
			subdomains <- sources.Subdomain{Source: source.Name(), Value: i}
		}
	}()

	return subdomains
}

func (source *Source) Name() string {
	return "threatcrowd"
}
