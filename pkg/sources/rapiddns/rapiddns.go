package rapiddns

import (
	"fmt"

	"github.com/signedsecurity/sigsubfind3r/pkg/sources"
)

type Source struct{}

func (source *Source) Run(domain string, session *sources.Session) chan sources.Subdomain {
	subdomains := make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		res, _ := session.SimpleGet(fmt.Sprintf("https://rapiddns.io/subdomain/%s?full=1", domain))

		for _, subdomain := range session.Extractor.FindAllString(string(res.Body()), -1) {
			subdomains <- sources.Subdomain{Source: source.Name(), Value: subdomain}
		}
	}()

	return subdomains
}

func (source *Source) Name() string {
	return "rapiddns"
}
