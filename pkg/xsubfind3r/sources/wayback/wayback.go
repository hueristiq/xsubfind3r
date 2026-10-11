package wayback

import (
	"context"
	"encoding/json"
	"fmt"

	hqgolimiter "github.com/hueristiq/hq-lib-limiter-go"
	"github.com/spf13/cast"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type Source struct{}

func (s *Source) Name() (name string) {
	name = sources.WAYBACK

	return
}

func (s *Source) Run(ctx context.Context, cfg *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		for page := uint(0); ; page++ {
			getURLsReqURL := "https://web.archive.org/cdx/search/cdx"

			if err := limiter.Wait(ctx); err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: s.Name(),
					Error:  fmt.Errorf("failed to wait on rate limiter: %w", err),
				}

				results <- result

				break
			}

			getURLsRes, err := cfg.Client.NewRequest().
				URL(getURLsReqURL).
				Param("url", "*."+domain+"/*").
				Param("output", "json").
				Param("collapse", "urlkey").
				Param("fl", "original").
				Param("pageSize", "100").
				Param("page", cast.ToString(page)).
				Do(ctx)
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

				sources.CloseBody(getURLsRes.Body)

				return
			}

			sources.CloseBody(getURLsRes.Body)

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

func (s *Source) NeedsKeys() (needs bool) {
	needs = false

	return
}

func (s *Source) UseKeys(keys ...string) {
}

var limiter = hqgolimiter.New(
	hqgolimiter.WithRequestsPerMinute(40),
)

var _ sources.Source = (*Source)(nil)

func New() (source sources.Source) {
	source = &Source{}

	return
}
