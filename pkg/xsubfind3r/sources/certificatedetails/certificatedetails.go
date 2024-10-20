package certificatedetails

import (
	"bufio"
	"fmt"

	"github.com/hueristiq/hq-go-http/status"
	"github.com/hueristiq/xsubfind3r/pkg/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type Source struct{}

func (source *Source) Run(cfg *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		getCertificateDetailsReqURL := fmt.Sprintf("https://certificatedetails.com/%s", domain)

		getCertificateDetailsRes, err := httpclient.SimpleGet(getCertificateDetailsReqURL)
		if err != nil && getCertificateDetailsRes.StatusCode != status.NotFound {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			httpclient.DiscardResponse(getCertificateDetailsRes)

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
