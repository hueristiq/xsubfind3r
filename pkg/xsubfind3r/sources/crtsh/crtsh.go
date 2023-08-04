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

func (source *Source) Run(_ *sources.Configuration, domain string) (subdomainsChannel chan sources.Subdomain) {
	subdomainsChannel = make(chan sources.Subdomain)

	go func() {
		defer close(subdomainsChannel)

		var err error

		getNameValuesReqURL := fmt.Sprintf("https://crt.sh/?q=%%25.%s&output=json", domain)

		var getNameValuesRes *fasthttp.Response

		getNameValuesRes, err = httpclient.SimpleGet(getNameValuesReqURL)
		if err != nil {
			return
		}

		var getNameValuesResData getNameValuesResponse

		if err := json.Unmarshal(getNameValuesRes.Body(), &getNameValuesResData); err != nil {
			return
		}

		var regex *regexp.Regexp

		regex, err = extractor.New(domain)
		if err != nil {
			return
		}

		for _, record := range getNameValuesResData {
			for _, value := range strings.Split(record.NameValue, "\n") {
				subdomain := regex.FindString(value)

				if subdomain == "" {
					continue
				}

				subdomainsChannel <- sources.Subdomain{Source: source.Name(), Value: subdomain}
			}
		}
	}()

	return
}

func (source *Source) Name() string {
	return "crtsh"
}
