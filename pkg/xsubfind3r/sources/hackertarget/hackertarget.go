package hackertarget

import (
	"bufio"
	"fmt"

	hqgohttp "github.com/hueristiq/hq-go-http"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type Source struct{}

func (s *Source) Name() (name string) {
	name = sources.HACKERTARGET

	return
}

func (s *Source) UseKeys(keys ...string) {
}

func (s *Source) Run(cfg *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		hostSearchReqURL := "https://api.hackertarget.com/hostsearch"
		hostSearchReqCFG := &hqgohttp.RequestConfiguration{
			Params: map[string]string{
				"q": domain,
			},
		}

		hostSearchRes, err := hqgohttp.Get(hostSearchReqURL, hostSearchReqCFG)
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  fmt.Errorf("request failed: %w", err),
			}

			results <- result

			return
		}

		scanner := bufio.NewScanner(hostSearchRes.Body)

		for scanner.Scan() {
			line := scanner.Text()

			if line == "" {
				continue
			}

			match := cfg.Extractor.FindAllString(line, -1)

			for _, subdomain := range match {
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

			hostSearchRes.Body.Close()

			return
		}

		hostSearchRes.Body.Close()
	}()

	return results
}

var _ sources.Source = (*Source)(nil)

func New() (source sources.Source) {
	source = &Source{}

	return
}
