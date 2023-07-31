package wayback

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"net/url"
	"regexp"
	"strings"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/extractor"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/valyala/fasthttp"
)

type Source struct{}

func (source *Source) Run(_ *sources.Configuration, domain string) (subdomainsChannel chan sources.Subdomain) {
	subdomainsChannel = make(chan sources.Subdomain)

	go func() {
		defer close(subdomainsChannel)

		var err error

		getPagesReqURL := fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=*.%s/*&output=txt&fl=original&collapse=urlkey&showNumPages=true", domain)

		var getPagesRes *fasthttp.Response

		getPagesRes, err = httpclient.SimpleGet(getPagesReqURL)
		if err != nil {
			return
		}

		var pages uint

		if err = json.Unmarshal(getPagesRes.Body(), &pages); err != nil {
			return
		}

		var regex *regexp.Regexp

		regex, err = extractor.New(domain)
		if err != nil {
			return
		}

		for page := uint(0); page < pages; page++ {
			getURLsReqURL := fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=*.%s/*&output=txt&fl=original&collapse=urlkey&page=%d", domain, page)

			var getURLsRes *fasthttp.Response

			getURLsRes, err = httpclient.SimpleGet(getURLsReqURL)
			if err != nil {
				return
			}

			scanner := bufio.NewScanner(bytes.NewReader(getURLsRes.Body()))

			for scanner.Scan() {
				line := scanner.Text()

				if line == "" {
					continue
				}

				line, _ = url.QueryUnescape(line)
				subdomain := regex.FindString(line)

				if subdomain != "" {
					subdomain = strings.ToLower(subdomain)
					subdomain = strings.TrimPrefix(subdomain, "25")
					subdomain = strings.TrimPrefix(subdomain, "2f")

					subdomainsChannel <- sources.Subdomain{Source: source.Name(), Value: subdomain}
				}
			}

			if err = scanner.Err(); err != nil {
				return
			}
		}
	}()

	return
}

func (source *Source) Name() string {
	return "wayback"
}
