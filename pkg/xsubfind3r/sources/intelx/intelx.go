package intelx

import (
	"encoding/json"
	"fmt"
	"time"

	hqurl "github.com/hueristiq/hqgoutils/url"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/valyala/fasthttp"
)

type searchResponseType struct {
	ID     string `json:"id"`
	Status int    `json:"status"`
}

type selectorType struct {
	Selectvalue string `json:"selectorvalue"`
}

type searchResultType struct {
	Selectors []selectorType `json:"selectors"`
	Status    int            `json:"status"`
}

type requestBody struct {
	Term       string        `json:"term"`
	Timeout    time.Duration `json:"timeout"`
	MaxResults int           `json:"maxresults"`
	Media      int           `json:"media"`
}

type Source struct{}

func (source *Source) Run(config *sources.Configuration) (subdomains chan sources.Subdomain) {
	subdomains = make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		var (
			err  error
			res  *fasthttp.Response
			body []byte
		)

		if config.Keys.IntelXKey == "" || config.Keys.IntelXHost == "" {
			return
		}

		searchURL := fmt.Sprintf("https://%s/phonebook/search?k=%s", config.Keys.IntelXHost, config.Keys.IntelXKey)
		reqBody := requestBody{
			Term:       config.Domain,
			MaxResults: 100000,
			Media:      0,
			Timeout:    20,
		}

		body, err = json.Marshal(reqBody)
		if err != nil {
			return
		}

		res, err = httpclient.SimplePost(searchURL, "application/json", body)
		if err != nil {
			return
		}

		var response searchResponseType

		if err = json.Unmarshal(res.Body(), &response); err != nil {
			return
		}

		resultsURL := fmt.Sprintf("https://%s/phonebook/search/result?k=%s&id=%s&limit=10000", config.Keys.IntelXHost, config.Keys.IntelXKey, response.ID)
		status := 0

		for status == 0 || status == 3 {
			res, err = httpclient.Get(resultsURL, "", nil)
			if err != nil {
				return
			}

			var response searchResultType

			if err = json.Unmarshal(res.Body(), &response); err != nil {
				return
			}

			status = response.Status

			for _, hostname := range response.Selectors {
				var parsedURL *hqurl.URL

				parsedURL, err = hqurl.Parse(hostname.Selectvalue)
				if err != nil {
					continue
				}

				subdomains <- sources.Subdomain{Source: source.Name(), Value: parsedURL.Domain}
			}
		}
	}()

	return
}

func (source *Source) Name() string {
	return "intelx"
}
