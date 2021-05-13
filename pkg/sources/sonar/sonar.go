package sonar

import (
	"encoding/json"
	"fmt"

	"github.com/signedsecurity/sigsubfind3r/pkg/sources"
)

type response struct {
	Subdomains []string
}

type Source struct{}

func (source *Source) Run(domain string, session *sources.Session) chan sources.Subdomain {
	subdomains := make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		res, _ := session.SimpleGet(fmt.Sprintf("https://sonar.omnisint.io/subdomains/%s", domain))

		var results []interface{}

		if err := json.Unmarshal(res.Body(), &results); err != nil {
			return
		}

		for _, i := range results {
			subdomains <- sources.Subdomain{Source: source.Name(), Value: i.(string)}
		}
	}()

	return subdomains
}

func (source *Source) Name() string {
	return "sonar"
}
