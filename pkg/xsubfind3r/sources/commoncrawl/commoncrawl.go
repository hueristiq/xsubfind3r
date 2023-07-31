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

type getIndexesResponse []struct {
	ID  string `json:"id"`
	API string `json:"cdx-API"`
}

type getURLsResponse struct {
	URL   string `json:"url"`
	Error string `json:"error"`
}

type Source struct{}

func (source *Source) Run(_ *sources.Configuration, domain string) (subdomainsChannel chan sources.Subdomain) {
	subdomainsChannel = make(chan sources.Subdomain)

	go func() {
		defer close(subdomainsChannel)

		getIndexesReqURL := "https://index.commoncrawl.org/collinfo.json"

		var err error

		var getIndexesRes *fasthttp.Response

		getIndexesRes, err = httpclient.SimpleGet(getIndexesReqURL)
		if err != nil {
			return
		}

		var getIndexesResData getIndexesResponse

		if err = json.Unmarshal(getIndexesRes.Body(), &getIndexesResData); err != nil {
			return
		}

		wg := new(sync.WaitGroup)

		for _, indexData := range getIndexesResData {
			wg.Add(1)

			go func(API string) {
				defer wg.Done()

				getURLsReqHeaders := map[string]string{
					"Host": "index.commoncrawl.org",
				}

				getURLsReqURL := fmt.Sprintf("%s?url=*.%s/*&output=json&fl=url", API, domain)

				var err error

				var getURLsRes *fasthttp.Response

				getURLsRes, err = httpclient.Get(getURLsReqURL, "", getURLsReqHeaders)
				if err != nil {
					return
				}

				scanner := bufio.NewScanner(bytes.NewReader(getURLsRes.Body()))

				for scanner.Scan() {
					var getURLsResData getURLsResponse

					if err = json.Unmarshal(scanner.Bytes(), &getURLsResData); err != nil {
						return
					}

					if getURLsResData.Error != "" {
						return
					}

					parsedURL, err := hqgourl.Parse(getURLsResData.URL)
					if err != nil {
						continue
					}

					subdomainsChannel <- sources.Subdomain{Source: source.Name(), Value: parsedURL.Domain}
				}

				if scanner.Err() != nil {
					return
				}
			}(indexData.API)
		}

		wg.Wait()
	}()

	return subdomainsChannel
}

func (source *Source) Name() string {
	return "commoncrawl"
}
