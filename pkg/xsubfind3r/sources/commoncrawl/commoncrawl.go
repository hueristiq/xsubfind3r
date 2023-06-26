package commoncrawl

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"sync"

	"github.com/hueristiq/hqgourl"
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
	ID       string `json:"id"`
	Name     string `json:"name"`
	TimeGate string `json:"timegate"`
	API      string `json:"cdx-api"`
}

func (source *Source) Run(config *sources.Configuration) (subdomains chan sources.Subdomain) {
	subdomains = make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		var (
			err error
			res *fasthttp.Response
		)

		indexURL := "https://index.commoncrawl.org/collinfo.json"

		res, err = httpclient.SimpleGet(indexURL)
		if err != nil {
			return
		}

		var indexes CommonAPIResult

		if err := json.Unmarshal(res.Body(), &indexes); err != nil {
			fmt.Println(err)
			return
		}

		wg := new(sync.WaitGroup)

		for _, u := range indexes {
			wg.Add(1)

			go func(api string) {
				defer wg.Done()

				var headers = map[string]string{"Host": "index.commoncrawl.org"}

				reqURL := fmt.Sprintf("%s?url=*.%s/*&output=json&fl=url", api, config.Domain)

				res, err := httpclient.Get(reqURL, "", headers)
				if err != nil {
					return
				}

				scanner := bufio.NewScanner(bytes.NewReader(res.Body()))

				for scanner.Scan() {
					var result CommonResult

					if err := json.Unmarshal(scanner.Bytes(), &result); err != nil {
						return
					}

					if result.Error != "" {
						continue
					}

					parsedURL, err := hqgourl.Parse(result.URL)
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
