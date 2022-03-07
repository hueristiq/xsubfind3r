package commoncrawl

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"

	"github.com/enenumxela/urlx/pkg/urlx"
	"github.com/signedsecurity/sigsubfind3r/pkg/sources"
)

type Source struct{}

type CommonPaginationResult struct {
	Blocks   uint `json:"blocks"`
	PageSize uint `json:"pageSize"`
	Pages    uint `json:"pages"`
}

type CommonResult struct {
	URL   string `json:"url"`
	Error string `json:"error"`
}

type CommonAPIResult []struct {
	ID  string `json:"id"`
	API string `json:"cdx-api"`
}

func (source *Source) Run(domain string, session *sources.Session) chan sources.Subdomain {
	subdomains := make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		res, err := session.SimpleGet("http://index.commoncrawl.org/collinfo.json")
		if err != nil {
			return
		}

		var apiRresults CommonAPIResult

		if err := json.Unmarshal(res.Body(), &apiRresults); err != nil {
			return
		}

		for _, u := range apiRresults {
			var headers = map[string]string{"Host": "index.commoncrawl.org"}

			res, err := session.Get(fmt.Sprintf("%s?url=*.%s/*&output=json&fl=url", u.API, domain), "", headers)
			if err != nil {
				continue
			}

			sc := bufio.NewScanner(bytes.NewReader(res.Body()))

			for sc.Scan() {
				var result CommonResult

				if err := json.Unmarshal(sc.Bytes(), &result); err != nil {
					return
				}

				if result.Error != "" {
					continue
				}

				URL, err := urlx.Parse(result.URL)
				if err != nil {
					continue
				}

				subdomains <- sources.Subdomain{Source: source.Name(), Value: URL.Domain}
			}
		}
	}()

	return subdomains
}

func (source *Source) Name() string {
	return "commoncrawl"
}
