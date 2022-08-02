package chaos

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/hqsubfind3r/pkg/sources"
	"github.com/valyala/fasthttp"
)

type response struct {
	Domain     string   `json:"domain"`
	Subdomains []string `json:"subdomains"`
	Count      int      `json:"count"`
}

type Source struct{}

func (source *Source) Run(domain string, session *sources.Session) chan sources.Subdomain {
	subdomains := make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		if session.Keys.Chaos == "" {
			return
		}

		res, err := session.Request(
			fasthttp.MethodGet,
			fmt.Sprintf("https://dns.projectdiscovery.io/dns/%s/subdomains", domain),
			"",
			map[string]string{"Authorization": session.Keys.Chaos},
			nil,
		)
		if err != nil {
			return
		}

		var results response

		if err := json.Unmarshal(res.Body(), &results); err != nil {
			return
		}

		for _, i := range results.Subdomains {
			subdomains <- sources.Subdomain{Source: source.Name(), Value: fmt.Sprintf("%s.%s", i, results.Domain)}
		}
	}()

	return subdomains
}

func (source *Source) Name() string {
	return "chaos"
}
