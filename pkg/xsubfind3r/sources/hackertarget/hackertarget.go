package hackertarget

import (
	"bufio"
	"bytes"
	"fmt"
	"net/url"
	"regexp"

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

		hostSearchReqURL := fmt.Sprintf("https://api.hackertarget.com/hostsearch/?q=%s", domain)

		var hostSearchRes *fasthttp.Response

		hostSearchRes, err = httpclient.SimpleGet(hostSearchReqURL)
		if err != nil {
			return
		}

		var regex *regexp.Regexp

		regex, err = extractor.New(domain)
		if err != nil {
			return
		}

		scanner := bufio.NewScanner(bytes.NewReader(hostSearchRes.Body()))

		for scanner.Scan() {
			line := scanner.Text()

			if line == "" {
				continue
			}

			line, _ = url.QueryUnescape(line)
			match := regex.FindAllString(line, -1)

			for _, subdomain := range match {
				subdomainsChannel <- sources.Subdomain{Source: source.Name(), Value: subdomain}
			}
		}

		if err = scanner.Err(); err != nil {
			return
		}
	}()

	return
}

func (source *Source) Name() string {
	return "hackertarget"
}
