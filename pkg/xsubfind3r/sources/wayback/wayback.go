package wayback

import (
	"encoding/json"
	"fmt"
	"net/http"

	hqgohttp "go.source.hueristiq.com/http"
	hqgolimiter "github.com/hueristiq/hq-go-limiter"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type Source struct{}

var limiter = hqgolimiter.New(&hqgolimiter.Configuration{
	RequestsPerMinute: 40,
})

func (source *Source) Run(cfg *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		var err error

		for page := uint(0); ; page++ {
			limiter.Wait()

			getURLsReqURL := fmt.Sprintf("https://web.archive.org/cdx/search/cdx?url=*.%s/*&output=json&collapse=urlkey&fl=original&pageSize=100&page=%d", domain, page)

			var getURLsRes *http.Response

			getURLsRes, err = hqgohttp.GET(getURLsReqURL).Send()
			if err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: source.Name(),
					Error:  err,
				}

				results <- result

				return
			}

			var getURLsResData [][]string

			if err = json.NewDecoder(getURLsRes.Body).Decode(&getURLsResData); err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: source.Name(),
					Error:  err,
				}

				results <- result

				getURLsRes.Body.Close()

				return
			}

			getURLsRes.Body.Close()

			// check if there's results, wayback's pagination response
			// is not always correct
			if len(getURLsResData) == 0 {
				break
			}

			// Slicing as [1:] to skip first result by default
			for _, entry := range getURLsResData[1:] {
				match := cfg.Extractor.FindAllString(entry[0], -1)

				for _, subdomain := range match {
					result := sources.Result{
						Type:   sources.ResultSubdomain,
						Source: source.Name(),
						Value:  subdomain,
					}

					results <- result
				}
			}
		}
	}()

	return results
}

func (source *Source) Name() string {
	return sources.WAYBACK
}
