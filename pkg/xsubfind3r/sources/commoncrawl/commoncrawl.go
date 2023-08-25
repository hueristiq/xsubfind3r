package commoncrawl

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/hueristiq/hqgourl"
	"github.com/hueristiq/xsubfind3r/pkg/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
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

func (source *Source) Run(_ *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		getIndexesReqURL := "https://index.commoncrawl.org/collinfo.json"

		var err error

		var getIndexesRes *http.Response

		getIndexesRes, err = httpclient.SimpleGet(getIndexesReqURL)
		if err != nil {
			result := sources.Result{
				Type:   sources.Error,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var getIndexesResData getIndexesResponse

		err = json.NewDecoder(getIndexesRes.Body).Decode(&getIndexesResData)
		if err != nil {
			result := sources.Result{
				Type:   sources.Error,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			getIndexesRes.Body.Close()

			return
		}

		getIndexesRes.Body.Close()

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

				var getURLsRes *http.Response

				getURLsRes, err = httpclient.Get(getURLsReqURL, "", getURLsReqHeaders)
				if err != nil {
					result := sources.Result{
						Type:   sources.Error,
						Source: source.Name(),
						Error:  err,
					}

					results <- result

					return
				}

				scanner := bufio.NewScanner(getURLsRes.Body)

				for scanner.Scan() {
					line := scanner.Text()
					if line == "" {
						continue
					}

					var getURLsResData getURLsResponse

					err = json.Unmarshal(scanner.Bytes(), &getURLsResData)
					if err != nil {
						result := sources.Result{
							Type:   sources.Error,
							Source: source.Name(),
							Error:  err,
						}

						results <- result

						continue
					}

					if getURLsResData.Error != "" {
						result := sources.Result{
							Type:   sources.Error,
							Source: source.Name(),
							Error:  fmt.Errorf("%s", getURLsResData.Error),
						}

						results <- result

						return
					}

					var parsedURL *hqgourl.URL

					parsedURL, err = hqgourl.Parse(getURLsResData.URL)
					if err != nil {
						result := sources.Result{
							Type:   sources.Error,
							Source: source.Name(),
							Error:  err,
						}

						results <- result

						continue
					}

					result := sources.Result{
						Type:   sources.Subdomain,
						Source: source.Name(),
						Value:  parsedURL.Domain,
					}

					results <- result
				}

				if err = scanner.Err(); err != nil {
					result := sources.Result{
						Type:   sources.Error,
						Source: source.Name(),
						Error:  err,
					}

					results <- result

					getURLsRes.Body.Close()

					return
				}

				getURLsRes.Body.Close()
			}(indexData.API)
		}

		wg.Wait()
	}()

	return results
}

func (source *Source) Name() string {
	return "commoncrawl"
}
