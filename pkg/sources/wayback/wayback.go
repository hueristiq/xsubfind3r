package wayback

import (
	"bufio"
	"bytes"
	"fmt"
	"net/url"

	"github.com/signedsecurity/sigsubfind3r/pkg/sources"
)

type Source struct{}

func (source *Source) Run(domain string, session *sources.Session) chan sources.Subdomain {
	subdomains := make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		res, _ := session.SimpleGet(fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=*.%s/*&output=txt&fl=original&collapse=urlkey", domain))

		scanner := bufio.NewScanner(bytes.NewReader(res.Body()))

		for scanner.Scan() {
			line := scanner.Text()

			if line == "" {
				continue
			}

			line, _ = url.QueryUnescape(line)
			subdomain := session.Extractor.FindString(line)

			subdomains <- sources.Subdomain{Source: source.Name(), Value: subdomain}
		}
	}()

	return subdomains
}

func (source *Source) Name() string {
	return "wayback"
}
