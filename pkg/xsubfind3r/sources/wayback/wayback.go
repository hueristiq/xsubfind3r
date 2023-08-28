package wayback

import (
	"encoding/json"
	"fmt"
	"net/http"
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
			getURLsReqURL := fmt.Sprintf("http://web.archive.org/cdx/search/cdx?url=*.%s/*&output=json&collapse=urlkey&fl=original&page=%d", domain, page)

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

			var getURLsResData [][]string

			err = json.NewDecoder(getURLsRes.Body).Decode(&getURLsResData)
			if err != nil {
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

			// check if there's results, wayback's pagination response
			// is not always correct when using a filter
			if len(getURLsResData) == 0 {
				break
			}

			// Slicing as [1:] to skip first result by default
			for _, entry := range getURLsResData[1:] {
				match := regex.FindAllString(entry[0], -1)

				for _, subdomain := range match {
					result := sources.Result{
						Type:   sources.Subdomain,
						Source: source.Name(),
						Value:  subdomain,
					}

					results <- result
				}
			}
		}
	}()

	return results
}

func (source *Source) Name() string {
	return "wayback"
}
