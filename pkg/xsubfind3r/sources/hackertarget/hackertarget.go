package hackertarget

import (
	"bufio"
	"fmt"

	hqgohttp "go.source.hueristiq.com/http"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type Source struct{}

func (source *Source) Run(cfg *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		hostSearchReqURL := fmt.Sprintf("https://api.hackertarget.com/hostsearch/?q=%s", domain)

		hostSearchRes, err := hqgohttp.GET(hostSearchReqURL).Send()
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
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
					Source: source.Name(),
					Value:  subdomain,
				}

				results <- result
			}
		}

		if err = scanner.Err(); err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			hostSearchRes.Body.Close()

			return
		}

		hostSearchRes.Body.Close()
	}()

	return results
}

func (source *Source) Name() string {
	return sources.HACKERTARGET
}
