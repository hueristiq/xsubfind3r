package chaos

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	hqgohttp "go.source.hueristiq.com/http"
	"go.source.hueristiq.com/http/method"
)

type getSubdomainsResponse struct {
	Domain     string   `json:"domain"`
	Subdomains []string `json:"subdomains"`
	Count      int      `json:"count"`
}

type Source struct{}

func (source *Source) Run(cfg *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		key, err := cfg.Keys.Chaos.PickRandom()
		if key == "" || err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		getSubdomainsReqURL := fmt.Sprintf("https://dns.projectdiscovery.io/dns/%s/subdomains", domain)

		getSubdomainsRes, err := hqgohttp.Request().Method(method.GET.String()).URL(getSubdomainsReqURL).AddHeader("Authorization", key).Send()
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var getSubdomainsResData getSubdomainsResponse

		if err = json.NewDecoder(getSubdomainsRes.Body).Decode(&getSubdomainsResData); err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			getSubdomainsRes.Body.Close()

			return
		}

		getSubdomainsRes.Body.Close()

		for _, subdomain := range getSubdomainsResData.Subdomains {
			result := sources.Result{
				Type:   sources.ResultSubdomain,
				Source: source.Name(),
				Value:  fmt.Sprintf("%s.%s", subdomain, getSubdomainsResData.Domain),
			}

			results <- result
		}
	}()

	return results
}

func (source *Source) Name() string {
	return sources.CHAOS
}
