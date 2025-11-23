package crtsh

import (
	"encoding/json"
	"fmt"
	"strings"

	hqgohttp "github.com/hueristiq/hq-go-http"
	hqgohttpheader "github.com/hueristiq/hq-go-http/header"
	hqgohttpmime "github.com/hueristiq/hq-go-http/mime"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type getNameValuesResponse []struct {
	ID        int    `json:"id"`
	NameValue string `json:"name_value"`
}

type Source struct{}

func (s *Source) Name() (name string) {
	name = sources.CRTSH

	return
}

func (s *Source) Run(_ *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		getNameValuesReqURL := "https://crt.sh"
		getNameValuesReqCFG := &hqgohttp.RequestConfiguration{
			Params: map[string]string{
				"q":      "%." + domain,
				"output": "json",
			},
			Headers: []hqgohttp.Header{
				hqgohttp.NewSetHeader(hqgohttpheader.ContentType.String(), hqgohttpmime.JSON.String()),
			},
		}

		getNameValuesRes, err := hqgohttp.Get(getNameValuesReqURL, getNameValuesReqCFG)
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  fmt.Errorf("request failed: %w", err),
			}

			results <- result

			return
		}

		var getNameValuesResData getNameValuesResponse

		if err = json.NewDecoder(getNameValuesRes.Body).Decode(&getNameValuesResData); err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  fmt.Errorf("failed to parse JSON response: %w", err),
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
					Source: s.Name(),
					Value:  subdomain,
				}

				results <- result
			}
		}
	}()

	return results
}

func (s *Source) NeedsKeys() (needs bool) {
	needs = false

	return
}

func (s *Source) UseKeys(keys ...string) {
}

var _ sources.Source = (*Source)(nil)

func New() (source sources.Source) {
	source = &Source{}

	return
}
