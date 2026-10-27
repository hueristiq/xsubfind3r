package rapiddns

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"strconv"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

var totalPattern = regexp.MustCompile(`Total:\s*<span[^>]*>(\d+)</span>`)

var rowPattern = regexp.MustCompile(`<th scope="row\s*">(\d+)</th>`)

type Source struct{}

func (s *Source) Name() (name string) {
	name = sources.RapidDNS

	return
}

func (s *Source) NeedsKeys() (needs bool) {
	needs = false

	return
}

func (s *Source) UseKeys(keys ...string) {
}

func (s *Source) Run(ctx context.Context, cfg *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		getSubdomainsReqURL := "https://rapiddns.io/subdomain/" + domain

		getSubdomainsRes, err := cfg.Client.Request().
			URL(getSubdomainsReqURL).
			Param("full", "1").
			Context(ctx).Do()
		if err != nil {
			result := sources.ErrorResult(s.Name(), fmt.Errorf("request failed: %w", err))

			results <- result

			return
		}

		body, err := io.ReadAll(getSubdomainsRes.Body)
		if err != nil {
			result := sources.ErrorResult(s.Name(), fmt.Errorf("failed to read response body: %w", err))

			results <- result

			_ = getSubdomainsRes.Body.Close()

			return
		}

		_ = getSubdomainsRes.Body.Close()

		src := string(body)

		for _, candidate := range sources.ExtractSubdomains(cfg.Extractor, src) {
			subdomain, valid := cfg.Validator(candidate)
			if !valid {
				continue
			}

			result := sources.SubdomainResult(s.Name(), subdomain)

			results <- result
		}

		// RapidDNS no longer renders pagination links and ignores the
		// page parameter, so guests receive a single capped page while
		// the markup still reports the full total.
		total := 0

		if match := totalPattern.FindStringSubmatch(src); len(match) > 1 {
			if parsed, err := strconv.Atoi(match[1]); err == nil {
				total = parsed
			}
		}

		rows := len(rowPattern.FindAllString(src, -1))

		if total > 0 && rows < total {
			result := sources.ErrorResult(s.Name(), fmt.Errorf("results truncated to %d of %d rows; authenticate for full data", rows, total))

			results <- result
		}
	}()

	return results
}

var _ sources.Source = (*Source)(nil)

func New() (source sources.Source) {
	source = &Source{}

	return
}
