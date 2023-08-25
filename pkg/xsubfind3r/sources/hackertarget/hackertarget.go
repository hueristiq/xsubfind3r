package hackertarget

import (
	"bufio"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/hueristiq/xsubfind3r/pkg/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/extractor"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
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

			line, _ = url.QueryUnescape(line)
			match := regex.FindAllString(line, -1)

			for _, subdomain := range match {
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
