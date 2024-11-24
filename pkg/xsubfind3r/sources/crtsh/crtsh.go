package crtsh

import (
	"encoding/json"
	"fmt"
	"strings"

	hqgohttp "github.com/hueristiq/hq-go-http"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type getNameValuesResponse []struct {
	ID        int    `json:"id"`
	NameValue string `json:"name_value"`
}

type Source struct{}

func (source *Source) Run(_ *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		getNameValuesReqURL := fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", domain)

		getNameValuesRes, err := hqgohttp.GET(getNameValuesReqURL).Send()
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var getNameValuesResData getNameValuesResponse

		if err = json.NewDecoder(getNameValuesRes.Body).Decode(&getNameValuesResData); err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			getNameValuesRes.Body.Close()

			return
		}

		getNameValuesRes.Body.Close()

		for _, record := range getNameValuesResData {
			subdomains := strings.Split(record.NameValue, "\n")

			for _, subdomain := range subdomains {
				if subdomain != domain && !strings.HasSuffix(subdomain, "."+domain) {
					continue
				}

				result := sources.Result{
					Type:   sources.ResultSubdomain,
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
	return sources.CRTSH
}
