package sublist3r

import (
	"encoding/json"
	"fmt"

	"github.com/signedsecurity/sigsubfind3r/pkg/sources"
)

type Source struct{}

func (source *Source) Run(domain string, session *sources.Session) chan sources.Subdomain {
	subdomains := make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		res, err := session.SimpleGet(fmt.Sprintf("https://api.sublist3r.com/search.php?domain=%s", domain))
		if err != nil {
			return
		}

		body := res.Body()

		var results []string

		if err := json.Unmarshal(body, &results); err != nil {
			return
		}

		for _, i := range results {
			subdomains <- sources.Subdomain{Source: source.Name(), Value: i}
		}
	}()

	return subdomains
}

func (source *Source) Name() string {
	return "sublist3r"
}
