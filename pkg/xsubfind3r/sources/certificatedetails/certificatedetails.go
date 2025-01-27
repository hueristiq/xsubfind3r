package certificatedetails

import (
	"bufio"
	"fmt"

	hqgohttp "go.source.hueristiq.com/http"
	"go.source.hueristiq.com/http/status"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type Source struct{}

func (source *Source) Run(cfg *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		getCertificateDetailsReqURL := fmt.Sprintf("https://certificatedetails.com/%s", domain)

		getCertificateDetailsRes, err := hqgohttp.GET(getCertificateDetailsReqURL).Send()
		if err != nil && getCertificateDetailsRes.StatusCode != status.NotFound.Int() {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		scanner := bufio.NewScanner(getCertificateDetailsRes.Body)

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

			getCertificateDetailsRes.Body.Close()

			return
		}

		getCertificateDetailsRes.Body.Close()
	}()

	return results
}

func (source *Source) Name() string {
	return sources.CERTIFICATEDETAILS
}
