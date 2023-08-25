package wayback

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

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

		getPagesReqURL := fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=*.%s/*&output=txt&fl=original&collapse=urlkey&showNumPages=true", domain)

		var getPagesRes *http.Response

		getPagesRes, err = httpclient.SimpleGet(getPagesReqURL)
		if err != nil {
			result := sources.Result{
				Type:   sources.Error,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var pages uint

		err = json.NewDecoder(getPagesRes.Body).Decode(&pages)
		if err != nil {
			result := sources.Result{
				Type:   sources.Error,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			getPagesRes.Body.Close()

			return
		}

		getPagesRes.Body.Close()

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

		for page := uint(0); page < pages; page++ {
			getURLsReqURL := fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=*.%s/*&output=txt&fl=original&collapse=urlkey&page=%d", domain, page)

			var getURLsRes *http.Response

			getURLsRes, err = httpclient.SimpleGet(getURLsReqURL)
			if err != nil {
				result := sources.Result{
					Type:   sources.Error,
					Source: source.Name(),
					Error:  err,
				}

				results <- result

				return
			}

			scanner := bufio.NewScanner(getURLsRes.Body)

			for scanner.Scan() {
				line := scanner.Text()

				if line == "" {
					continue
				}

				line, _ = url.QueryUnescape(line)
				subdomain := regex.FindString(line)

				if subdomain != "" {
					subdomain = strings.ToLower(subdomain)
					subdomain = strings.TrimPrefix(subdomain, "25")
					subdomain = strings.TrimPrefix(subdomain, "2f")

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

				getURLsRes.Body.Close()

				return
			}

			getURLsRes.Body.Close()
		}
	}()

	return results
}

func (source *Source) Name() string {
	return "wayback"
}
