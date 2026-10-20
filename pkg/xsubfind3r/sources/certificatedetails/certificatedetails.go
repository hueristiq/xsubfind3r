package certificatedetails

import (
	"bufio"
	"context"
	"fmt"

	hqgohttpstatus "github.com/hueristiq/hq-lib-http-go/status"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type Source struct{}

func (s *Source) Name() (name string) {
	name = sources.Certificatedetails

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

		getCertificateDetailsReqURL := "https://certificatedetails.com/" + domain

		getCertificateDetailsRes, err := cfg.Client.NewRequest().
			URL(getCertificateDetailsReqURL).
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

		if getCertificateDetailsRes.StatusCode == hqgohttpstatus.NotFound.Int() {
			_ = getCertificateDetailsRes.Body.Close()

			return
		}

		scanner := bufio.NewScanner(getCertificateDetailsRes.Body)

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

				result := sources.Result{
					Type:   sources.ResultSubdomain,
					Source: s.Name(),
					Value:  subdomain,
				}

				results <- result
			}
		}

		if err = scanner.Err(); err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  fmt.Errorf("failed to read response body: %w", err),
			}

			results <- result

			_ = getCertificateDetailsRes.Body.Close()

			return
		}

		_ = getCertificateDetailsRes.Body.Close()
	}()

	return results
}

var _ sources.Source = (*Source)(nil)

func New() (source sources.Source) {
	source = &Source{}

	return
}
