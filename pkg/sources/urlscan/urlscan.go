package urlscan

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/subfind3r/pkg/sources"
)

type response struct {
	Results []struct {
		Page struct {
			Domain string `json:"domain"`
		} `json:"page"`
	} `json:"results"`
}

type Source struct{}

func (source *Source) Run(domain string, session *sources.Session) chan sources.Subdomain {
	subdomains := make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		res, err := session.SimpleGet(fmt.Sprintf("https://urlscan.io/api/v1/search/?q=domain:%s", domain))
		if err != nil {
			return
		}

		body := res.Body()

		var results response

		if err := json.Unmarshal(body, &results); err != nil {
			return
		}

		for _, i := range results.Results {
			subdomains <- sources.Subdomain{Source: source.Name(), Value: i.Page.Domain}
		}
	}()

	return subdomains
}

func (source *Source) Name() string {
	return "urlscan"
}
