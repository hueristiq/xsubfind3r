package crtsh

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/hueristiq/hqsubfind3r/pkg/sources"
)

type response struct {
	ID        int    `json:"id"`
	NameValue string `json:"name_value"`
}

type Source struct{}

func (source *Source) Run(domain string, session *sources.Session) chan sources.Subdomain {
	subdomains := make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		res, err := session.SimpleGet(fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", domain))
		if err != nil {
			return
		}

		var results []response

		if err := json.Unmarshal(res.Body(), &results); err != nil {
			return
		}

		for _, i := range results {
			x := strings.Split(i.NameValue, "\n")

			for _, j := range x {
				subdomains <- sources.Subdomain{Source: source.Name(), Value: j}
			}

		}
	}()

	return subdomains
}

func (source *Source) Name() string {
	return "crtsh"
}
