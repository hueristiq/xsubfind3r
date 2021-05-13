package hackertarget

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

		res, _ := session.SimpleGet(fmt.Sprintf("http://api.hackertarget.com/hostsearch/?q=%s", domain))

		scanner := bufio.NewScanner(bytes.NewReader(res.Body()))

		for scanner.Scan() {
			line := scanner.Text()

			if line == "" {
				continue
			}

			line, _ = url.QueryUnescape(line)
			match := session.Extractor.FindAllString(line, -1)

			for _, subdomain := range match {
				subdomains <- sources.Subdomain{Source: source.Name(), Value: subdomain}
			}
		}
	}()

	return subdomains
}

func (source *Source) Name() string {
	return "hackertarget"
}
