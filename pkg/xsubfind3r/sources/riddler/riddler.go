package riddler

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

		res, err = httpclient.SimpleGet(fmt.Sprintf("https://riddler.io/search?q=pld:%s&view_type=data_table", config.Domain))
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
			match := config.SubdomainsRegex.FindAllString(line, -1)

			for _, subdomain := range match {
				subdomains <- sources.Subdomain{Source: source.Name(), Value: subdomain}
			}
		}

		if err = scanner.Err(); err != nil {
			return
		}
	}()

	return
}

func (source *Source) Name() string {
	return "riddler"
}
