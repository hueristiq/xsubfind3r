package ximcx

import (
	"encoding/json"
	"fmt"

	"github.com/signedsecurity/sigsubfind3r/pkg/sources"
)

type Source struct{}

type response struct {
	Code    int64  `json:"code"`
	Message string `json:"message"`
	Data    []struct {
		Domain string `json:"domain"`
	} `json:"data"`
}

func (source *Source) Run(domain string, session *sources.Session) chan sources.Subdomain {
	subdomains := make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		res, _ := session.SimpleGet(fmt.Sprintf("http://sbd.ximcx.cn/DomainServlet?domain=%s", domain))

		var results response

		if err := json.Unmarshal(res.Body(), &results); err != nil {
			return
		}

		for _, result := range results.Data {
			subdomains <- sources.Subdomain{Source: source.Name(), Value: result.Domain}
		}
	}()

	return subdomains
}

func (source *Source) Name() string {
	return "ximcx"
}
