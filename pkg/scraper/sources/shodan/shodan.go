package shodan

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hueristiq/xsubfind3r/pkg/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/scraper/sources"
)

type getDNSResponse struct {
	Domain     string   `json:"domain"`
	Subdomains []string `json:"subdomains"`
	Result     int      `json:"result"`
	Error      string   `json:"error"`
}

type Source struct{}

func (source *Source) Run(config *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		var err error

		var key string

		key, err = sources.PickRandom(config.Keys.Shodan)
		if key == "" || err != nil {
			result := sources.Result{
				Type:   sources.Error,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		getDNSReqURL := fmt.Sprintf("https://api.shodan.io/dns/domain/%s?key=%s", domain, key)

		var getDNSRes *http.Response

		getDNSRes, err = httpclient.SimpleGet(getDNSReqURL)
		if err != nil {
			result := sources.Result{
				Type:   sources.Error,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var getDNSResData getDNSResponse

		err = json.NewDecoder(getDNSRes.Body).Decode(&getDNSResData)
		if err != nil {
			result := sources.Result{
				Type:   sources.Error,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			getDNSRes.Body.Close()

			return
		}

		getDNSRes.Body.Close()

		for index := range getDNSResData.Subdomains {
			subdomain := getDNSResData.Subdomains[index]

			result := sources.Result{
				Type:   sources.Subdomain,
				Source: source.Name(),
				Value:  fmt.Sprintf("%s.%s", subdomain, domain),
			}

			results <- result
		}
	}()

	return results
}

func (source *Source) Name() string {
	return "shodan"
}
