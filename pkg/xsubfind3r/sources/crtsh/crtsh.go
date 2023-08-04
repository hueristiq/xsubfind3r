package crtsh

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/extractor"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/valyala/fasthttp"
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

		var err error

		getNameValuesReqURL := fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", domain)

		var getNameValuesRes *fasthttp.Response

		getNameValuesRes, err = httpclient.SimpleGet(getNameValuesReqURL)
		if err != nil {
			result := sources.Result{
				Type:   sources.Error,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var getNameValuesResData getNameValuesResponse

		err = json.Unmarshal(getNameValuesRes.Body(), &getNameValuesResData)
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

		for _, record := range getNameValuesResData {
			for _, value := range strings.Split(record.NameValue, "\n") {
				subdomain := regex.FindString(value)

				if subdomain == "" {
					continue
				}

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
	return "crtsh"
}
