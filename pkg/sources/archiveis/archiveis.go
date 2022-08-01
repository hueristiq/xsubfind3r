package archiveis

import (
	"fmt"

	"github.com/hueristiq/subfind3r/pkg/sources"
)

type Source struct{}

func (source *Source) Run(domain string, session *sources.Session) chan sources.Subdomain {
	subdomains := make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		res, err := session.SimpleGet(fmt.Sprintf("https://archive.is/*.%s", domain))
		if err != nil {
			return
		}

		src := string(res.Body())

		for _, subdomain := range session.Extractor.FindAllString(src, -1) {
			subdomains <- sources.Subdomain{Source: source.Name(), Value: subdomain}
		}
	}()

	return subdomains
}

func (s *Source) Name() string {
	return "archiveis"
}
