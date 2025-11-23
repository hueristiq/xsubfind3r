package wayback

import (
	"encoding/json"
	"fmt"

	hqgohttp "github.com/hueristiq/hq-go-http"
	hqgolimiter "github.com/hueristiq/hq-go-limiter"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/spf13/cast"
)

type Source struct{}

func (s *Source) Name() (name string) {
	name = sources.WAYBACK

	return
}

func (s *Source) Run(cfg *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		for page := uint(0); ; page++ {
			limiter.Wait()

			getURLsReqURL := "https://web.archive.org/cdx/search/cdx"
			getURLsReqCFG := &hqgohttp.RequestConfiguration{
				Params: map[string]string{
					"url":      "*." + domain + "/*",
					"output":   "json",
					"collapse": "urlkey",
					"fl":       "original",
					"pageSize": "100",
					"page":     cast.ToString(page),
				},
			}

			getURLsRes, err := hqgohttp.Get(getURLsReqURL, getURLsReqCFG)
			if err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: s.Name(),
					Error:  fmt.Errorf("request failed: %w", err),
				}

				results <- result

				return
			}

			var getURLsResData [][]string

			if err = json.NewDecoder(getURLsRes.Body).Decode(&getURLsResData); err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: s.Name(),
					Error:  fmt.Errorf("failed to parse JSON response: %w", err),
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
						Source: s.Name(),
						Value:  subdomain,
					}

					results <- result
				}
			}
		}
	}()

	return results
}

func (s *Source) UseKeys(keys ...string) {
}

var limiter = hqgolimiter.New(&hqgolimiter.Configuration{
	RequestsPerMinute: 40,
})

var _ sources.Source = (*Source)(nil)

func New() (source sources.Source) {
	source = &Source{}

	return
}
