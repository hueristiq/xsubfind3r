package cebaidu

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/subfind3r/pkg/sources"
)

type response struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    []struct {
		Domain string `json:"domain"`
	} `json:"data"`
}

type Source struct{}

func (source *Source) Run(domain string, session *sources.Session) chan sources.Subdomain {
	subdomains := make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		res, err := session.SimpleGet(fmt.Sprintf("https://ce.baidu.com/index/getRelatedSites?site_address=%s", domain))
		if err != nil {
			return
		}

		var results response

		if err := json.Unmarshal(res.Body(), &results); err != nil {
			return
		}

		for _, i := range results.Data {
			subdomains <- sources.Subdomain{Source: source.Name(), Value: i.Domain}
		}
	}()

	return subdomains
}

func (source *Source) Name() string {
	return "cebaidu"
}
