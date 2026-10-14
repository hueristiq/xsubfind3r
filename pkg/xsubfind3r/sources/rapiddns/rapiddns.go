package rapiddns

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strconv"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

var pagePattern = regexp.MustCompile(`class="page-link" href="/subdomain/[^"]+\?page=(\d+)">`)

type Source struct{}

func (s *Source) Name() (name string) {
	name = sources.RAPIDDNS

	return
}

func (s *Source) Run(ctx context.Context, cfg *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		page := 1
		maxPages := 1

		for {
			select {
			case <-ctx.Done():
				return
			default:
			}

			getSubdomainsReqURL := "https://rapiddns.io/subdomain/" + domain

			getSubdomainsRes, err := cfg.Client.NewRequest().
				URL(getSubdomainsReqURL).
				Param("full", "1").
				Param("page", strconv.Itoa(page)).
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

			body, err := io.ReadAll(getSubdomainsRes.Body)
			if err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: s.Name(),
					Error:  fmt.Errorf("failed to read response body: %w", err),
				}

				results <- result

				sources.CloseBody(getSubdomainsRes.Body)

				return
			}

			sources.CloseBody(getSubdomainsRes.Body)

			src := string(body)

			subdomains := cfg.Extract(src)

			for _, subdomain := range subdomains {
				result := sources.Result{
					Type:   sources.ResultSubdomain,
					Source: s.Name(),
					Value:  subdomain,
				}

				results <- result
			}

			if maxPages == 1 {
				maxPages = paginationCount(src)
			}

			if page >= maxPages {
				break
			}

			page++
		}
	}()

	return results
}

func paginationCount(src string) (maxPages int) {
	maxPages = 1

	matches := pagePattern.FindAllStringSubmatch(src, -1)

	if len(matches) == 0 {
		return
	}

	lastMatch := matches[len(matches)-1]

	if len(lastMatch) < 2 {
		return
	}

	if parsed, err := strconv.Atoi(lastMatch[1]); err == nil {
		maxPages = parsed
	}

	return
}

func (s *Source) NeedsKeys() (needs bool) {
	needs = false

	return
}

func (s *Source) UseKeys(keys ...string) {
}

var _ sources.Source = (*Source)(nil)

func New() (source sources.Source) {
	source = &Source{}

	return
}
