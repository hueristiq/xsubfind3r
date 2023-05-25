package commoncrawl

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	hqurl "github.com/hueristiq/hqgoutils/url"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/valyala/fasthttp"
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

func (source *Source) Run(config *sources.Configuration) (subdomains chan sources.Subdomain) {
	subdomains = make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		var (
			err error
			res *fasthttp.Response
		)

		res, err = httpclient.SimpleGet("https://index.commoncrawl.org/collinfo.json")
		if err != nil {
			return
		}

		var apiRresults CommonAPIResult

		if err := json.Unmarshal(res.Body(), &apiRresults); err != nil {
			fmt.Println(err)
			return
		}

		wg := new(sync.WaitGroup)

		for _, u := range apiRresults {
			wg.Add(1)

			go func(api string) {
				defer wg.Done()

				var headers = map[string]string{"Host": "index.commoncrawl.org"}

				res, err := httpclient.Get(fmt.Sprintf("%s?url=*.%s/*&output=json&fl=url", api, config.Domain), "", headers)
				if err != nil {
					return
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

					parsedURL, err := hqurl.Parse(result.URL)
					if err != nil {
						continue
					}

					subdomains <- sources.Subdomain{Source: source.Name(), Value: parsedURL.Domain}
				}
			}(u.API)
		}

		wg.Wait()
	}()

	return subdomains
}

func (source *Source) Name() string {
	return "commoncrawl"
}
