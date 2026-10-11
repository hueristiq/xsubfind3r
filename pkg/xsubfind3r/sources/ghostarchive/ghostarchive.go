package ghostarchive

import (
	"context"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"

	hqgohttpstatus "github.com/hueristiq/hq-lib-http-go/status"
	hqgolimiter "github.com/hueristiq/hq-lib-limiter-go"
	"github.com/spf13/cast"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

var resultPattern = regexp.MustCompile(`<a href="/archive/[^"]*">([^<]+)</a>`)

type Source struct{}

func (s *Source) Name() (name string) {
	name = sources.GHOSTARCHIVE

	return
}

func (s *Source) Run(ctx context.Context, cfg *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		term := domain

		if strings.Count(domain, ".") == 1 {
			term = "." + domain
		}

		for page := uint(0); ; page++ {
			if err := limiter.Wait(ctx); err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: s.Name(),
					Error:  fmt.Errorf("failed to wait on rate limiter: %w", err),
				}

				results <- result

				break
			}

			getSubdomainsReqURL := "https://ghostarchive.org/search"

			getSubdomainsRes, err := cfg.Client.NewRequest().
				URL(getSubdomainsReqURL).
				Param("term", term).
				Param("page", cast.ToString(page)).
				Do(ctx)
			if err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: s.Name(),
					Error:  fmt.Errorf("request failed: %w", err),
				}

				results <- result

				break
			}

			status := getSubdomainsRes.StatusCode

			body, err := io.ReadAll(getSubdomainsRes.Body)
			if err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: s.Name(),
					Error:  fmt.Errorf("failed to read response body: %w", err),
				}

				results <- result

				sources.CloseBody(getSubdomainsRes.Body)

				break
			}

			sources.CloseBody(getSubdomainsRes.Body)

			if status == hqgohttpstatus.TooManyRequests.Int() || status == hqgohttpstatus.ServiceUnavailable.Int() {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: s.Name(),
					Error:  fmt.Errorf("request failed with status: %v", status),
				}

				results <- result

				break
			}

			src := string(body)

			if strings.Contains(src, "The site is under maintenance") || strings.Contains(src, "No archives for that site") {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: s.Name(),
					Error:  errors.New("no archives for that site"),
				}

				results <- result

				break
			}

			matches := resultPattern.FindAllStringSubmatch(src, -1)

			if len(matches) == 0 {
				break
			}

			for _, match := range matches {
				subdomains := cfg.Extractor.FindAllString(match[1], -1)

				for _, subdomain := range subdomains {
					result := sources.Result{
						Type:   sources.ResultSubdomain,
						Source: s.Name(),
						Value:  subdomain,
					}

					results <- result
				}
			}

			if !strings.Contains(src, "Next Page") && !strings.Contains(src, ">»</a>") {
				break
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
	hqgolimiter.WithRequestsPerMinute(30),
)

var _ sources.Source = (*Source)(nil)

func New() (source sources.Source) {
	source = &Source{}

	return
}
