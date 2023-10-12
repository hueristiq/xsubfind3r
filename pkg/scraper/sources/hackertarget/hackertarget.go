package hackertarget

import (
	"bufio"
	"fmt"
	"net/http"
	"regexp"

	"github.com/hueristiq/xsubfind3r/pkg/extractor"
	"github.com/hueristiq/xsubfind3r/pkg/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/scraper/sources"
)

type Source struct{}

func (source *Source) Run(_ *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		var err error

		hostSearchReqURL := fmt.Sprintf("https://api.hackertarget.com/hostsearch/?q=%s", domain)

		var hostSearchRes *http.Response

		hostSearchRes, err = httpclient.SimpleGet(hostSearchReqURL)
		if err != nil {
			result := sources.Result{
				Type:   sources.Error,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var regex *regexp.Regexp

		regex, err = extractor.New(domain)
		if err != nil {
			result := sources.Result{
				Type:   sources.Error,
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

			match := regex.FindAllString(line, -1)

			for index := range match {
				subdomain := match[index]

				result := sources.Result{
					Type:   sources.Subdomain,
					Source: source.Name(),
					Value:  subdomain,
				}

				results <- result
			}
		}

		if err = scanner.Err(); err != nil {
			result := sources.Result{
				Type:   sources.Error,
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
	return "hackertarget"
}
