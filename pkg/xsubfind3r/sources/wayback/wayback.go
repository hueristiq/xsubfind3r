package wayback

import (
	"bufio"
	"bytes"
	"fmt"
	"net/url"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/valyala/fasthttp"
)

type Source struct{}

func (source *Source) Run(config *sources.Configuration) (subdomains chan sources.Subdomain) {
	subdomains = make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		var (
			err error
			res *fasthttp.Response
		)

		res, err = httpclient.SimpleGet(fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=*.%s/*&output=txt&fl=original&collapse=urlkey", config.Domain))
		if err != nil {
			return
		}

		scanner := bufio.NewScanner(bytes.NewReader(res.Body()))

		for scanner.Scan() {
			line := scanner.Text()

			if line == "" {
				continue
			}

			line, _ = url.QueryUnescape(line)
			subdomain := config.SubdomainsRegex.FindString(line)

			subdomains <- sources.Subdomain{Source: source.Name(), Value: subdomain}
		}

		if err = scanner.Err(); err != nil {
			return
		}
	}()

	return
}

func (source *Source) Name() string {
	return "wayback"
}
