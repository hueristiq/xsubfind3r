package hackertarget

import (
	"bufio"
	"context"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type Source struct{}

func (s *Source) Name() (name string) {
	name = sources.Hackertarget

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

		hostSearchReqURL := "https://api.hackertarget.com/hostsearch"

		hostSearchRes, err := cfg.Client.Request().
			URL(hostSearchReqURL).
			Param("q", domain).
			Context(ctx).Do()
		if err != nil {
			result := sources.ErrorResult(s.Name(), fmt.Errorf("request failed: %w", err))

			results <- result

			return
		}

		scanner := bufio.NewScanner(hostSearchRes.Body)

		for scanner.Scan() {
			line := scanner.Text()

			if line == "" {
				continue
			}

			for _, candidate := range sources.ExtractSubdomains(cfg.Extractor, line) {
				subdomain, valid := cfg.Validator(candidate)
				if !valid {
					continue
				}

				result := sources.SubdomainResult(s.Name(), subdomain)

				results <- result
			}
		}

		if err = scanner.Err(); err != nil {
			result := sources.ErrorResult(s.Name(), fmt.Errorf("failed to read response body: %w", err))

			results <- result

			_ = hostSearchRes.Body.Close()

			return
		}

		_ = hostSearchRes.Body.Close()
	}()

	return results
}

var _ sources.Source = (*Source)(nil)

func New() (source sources.Source) {
	source = &Source{}

	return
}
