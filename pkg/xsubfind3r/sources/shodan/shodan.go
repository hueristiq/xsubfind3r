package shodan

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	hqgohttp "go.source.hueristiq.com/http"
	"go.source.hueristiq.com/http/method"
)

type getDNSResponse struct {
	Domain     string   `json:"domain"`
	Subdomains []string `json:"subdomains"`
	Result     int      `json:"result"`
	Error      string   `json:"error"`
}

type Source struct{}

func (source *Source) Run(cfg *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		key, err := cfg.Keys.Shodan.PickRandom()
		if key == "" || err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		getDNSReqURL := fmt.Sprintf("https://api.shodan.io/dns/domain/%s?key=%s", domain, key)

		var getDNSRes *http.Response

		getDNSRes, err = hqgohttp.Request().Method(method.GET.String()).URL(getDNSReqURL).Send()
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var getDNSResData getDNSResponse

		if err = json.NewDecoder(getDNSRes.Body).Decode(&getDNSResData); err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			getDNSRes.Body.Close()

			return
		}

		getDNSRes.Body.Close()

		for _, subdomain := range getDNSResData.Subdomains {
			result := sources.Result{
				Type:   sources.ResultSubdomain,
				Source: source.Name(),
				Value:  fmt.Sprintf("%s.%s", subdomain, domain),
			}

			results <- result
		}
	}()

	return results
}

func (source *Source) Name() string {
	return sources.SHODAN
}
