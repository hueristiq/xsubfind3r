package arquivo

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type cdxRecord struct {
	URL string `json:"url"`
}

type Source struct{}

func (s *Source) Name() (name string) {
	name = sources.Arquivo

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

		getSubdomainsReqURL := "https://arquivo.pt/wayback/cdx"

		getSubdomainsRes, err := cfg.Client.NewRequest().
			URL(getSubdomainsReqURL).
			Param("url", "*."+domain+"/*").
			Param("output", "json").
			Param("fl", "url").
			Param("collapse", "urlkey").
			Param("limit", "100000").
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

		scanner := bufio.NewScanner(getSubdomainsRes.Body)

		for scanner.Scan() {
			line := scanner.Text()

			if line == "" {
				continue
			}

			var record cdxRecord

			if err = json.Unmarshal([]byte(line), &record); err != nil {
				continue
			}

			subdomains := cfg.Extract(record.URL)

			for _, subdomain := range subdomains {
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

			_ = getSubdomainsRes.Body.Close()

			return
		}

		_ = getSubdomainsRes.Body.Close()
	}()

	return results
}

var _ sources.Source = (*Source)(nil)

func New() (source sources.Source) {
	source = &Source{}

	return
}
