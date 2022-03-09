package intelx

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/enenumxela/urlx/pkg/urlx"
	"github.com/signedsecurity/sigsubfind3r/pkg/sources"
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

func (source *Source) Run(domain string, session *sources.Session) chan sources.Subdomain {
	subdomains := make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		if session.Keys.IntelXKey == "" || session.Keys.IntelXHost == "" {
			return
		}

		searchURL := fmt.Sprintf("https://%s/phonebook/search?k=%s", session.Keys.IntelXHost, session.Keys.IntelXKey)
		reqBody := requestBody{
			Term:       domain,
			MaxResults: 100000,
			Media:      0,
			Timeout:    20,
		}

		body, err := json.Marshal(reqBody)
		if err != nil {
			return
		}

		res, err := session.SimplePost(searchURL, "application/json", body)
		if err != nil {
			return
		}

		var response searchResponseType

		if err := json.Unmarshal(res.Body(), &response); err != nil {
			return
		}

		resultsURL := fmt.Sprintf("https://%s/phonebook/search/result?k=%s&id=%s&limit=10000", session.Keys.IntelXHost, session.Keys.IntelXKey, response.ID)
		status := 0
		for status == 0 || status == 3 {
			res, err = session.Get(resultsURL, "", nil)
			if err != nil {
				return
			}

			var response searchResultType

			if err := json.Unmarshal(res.Body(), &response); err != nil {
				return
			}

			status = response.Status
			for _, hostname := range response.Selectors {
				URL, err := urlx.Parse(hostname.Selectvalue)
				if err != nil {
					continue
				}

				subdomains <- sources.Subdomain{Source: source.Name(), Value: URL.Domain}
			}
		}
	}()

	return subdomains
}

func (source *Source) Name() string {
	return "intelx"
}
