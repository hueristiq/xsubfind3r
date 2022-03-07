package archiveis

import (
	"fmt"

	"github.com/signedsecurity/sigsubfind3r/pkg/sources"
)

type Source struct{}

func (source *Source) Run(domain string, session *sources.Session) chan sources.Subdomain {
	subdomains := make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		res, _ := session.SimpleGet(fmt.Sprintf("https://archive.is/*.%s", domain))

		// resp, err := session.SimpleGet(ctx, fmt.Sprintf("https://archive.is/*.%s", domain))

		// if err != nil {
		// 	results <- subscraping.Result{Source: "archiveis", Type: subscraping.Error, Error: err}
		// 	session.DiscardHTTPResponse(resp)
		// 	return
		// }

		// body, err := ioutil.ReadAll(resp.Body)
		// if err != nil {
		// 	results <- subscraping.Result{Source: "archiveis", Type: subscraping.Error, Error: err}
		// 	resp.Body.Close()
		// 	return
		// }

		// resp.Body.Close()

		src := string(res.Body())

		for _, subdomain := range session.Extractor.FindAllString(src, -1) {
			// results <- subscraping.Result{Source: "archiveis", Type: subscraping.Subdomain, Value: subdomain}
			subdomains <- sources.Subdomain{Source: source.Name(), Value: subdomain}
		}
	}()

	return subdomains
}

func (s *Source) Name() string {
	return "archiveis"
}
