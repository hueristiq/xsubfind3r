package bufferover

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/hueristiq/xsubfind3r/pkg/extractor"
	"github.com/hueristiq/xsubfind3r/pkg/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/scraper/sources"
)

type getTLSLogsSearchResponse struct {
	Meta struct {
		Errors []string `json:"Errors"`
	} `json:"Meta"`
	FDNSA   []string `json:"FDNS_A"`
	RDNS    []string `json:"RDNS"`
	Results []string `json:"Results"`
}

type Source struct{}

func (source *Source) Run(config *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		var err error

		var key string

		key, err = sources.PickRandom(config.Keys.Bufferover)
		if key == "" || err != nil {
			result := sources.Result{
				Type:   sources.Error,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		getTLSLogsSearchReqHeaders := map[string]string{"x-api-key": key}

		getTLSLogsSearchReqURL := fmt.Sprintf("https://tls.bufferover.run/dns?q=.%s", domain)

		var getTLSLogsSearchRes *http.Response

		getTLSLogsSearchRes, err = httpclient.Get(getTLSLogsSearchReqURL, "", getTLSLogsSearchReqHeaders)
		if err != nil {
			result := sources.Result{
				Type:   sources.Error,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var getTLSLogsSearchResData getTLSLogsSearchResponse

		if err = json.NewDecoder(getTLSLogsSearchRes.Body).Decode(&getTLSLogsSearchResData); err != nil {
			result := sources.Result{
				Type:   sources.Error,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			getTLSLogsSearchRes.Body.Close()

			return
		}

		getTLSLogsSearchRes.Body.Close()

		if len(getTLSLogsSearchResData.Meta.Errors) > 0 {
			result := sources.Result{
				Type:   sources.Error,
				Source: source.Name(),
				Error:  fmt.Errorf("%s", strings.Join(getTLSLogsSearchResData.Meta.Errors, ", ")),
			}

			results <- result
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

		var entries []string

		if len(getTLSLogsSearchResData.FDNSA) > 0 {
			entries = getTLSLogsSearchResData.FDNSA
			entries = append(entries, getTLSLogsSearchResData.RDNS...)
		} else if len(getTLSLogsSearchResData.Results) > 0 {
			entries = getTLSLogsSearchResData.Results
		}

		for _, entry := range entries {
			subdomains := regex.FindAllString(entry, -1)

			for _, subdomain := range subdomains {
				result := sources.Result{
					Type:   sources.Subdomain,
					Source: source.Name(),
					Value:  subdomain,
				}

				results <- result
			}
		}
	}()

	return results
}

func (source *Source) Name() string {
	return "bufferover"
}
