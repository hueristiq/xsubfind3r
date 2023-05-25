package bufferover

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/valyala/fasthttp"
)

type response struct {
	Meta struct {
		Errors []string `json:"Errors"`
	} `json:"Meta"`
	FDNSA   []string `json:"FDNS_A"`
	RDNS    []string `json:"RDNS"`
	Results []string `json:"Results"`
}

type Source struct{}

func (source *Source) Run(config *sources.Configuration) (subdomains chan sources.Subdomain) {
	subdomains = make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		// Run enumeration on subdomain dataset for historical SONAR datasets
		source.getData(config, fmt.Sprintf("https://dns.bufferover.run/dns?q=.%s", config.Domain), subdomains)
		source.getData(config, fmt.Sprintf("https://tls.bufferover.run/dns?q=.%s", config.Domain), subdomains)
	}()

	return subdomains
}

func (source *Source) getData(config *sources.Configuration, sourceURL string, subdomains chan sources.Subdomain) {
	var (
		err error
		res *fasthttp.Response
	)

	res, err = httpclient.SimpleGet(sourceURL)
	if err != nil {
		return
	}

	var results response

	if err = json.Unmarshal(res.Body(), &results); err != nil {
		return
	}

	var subs []string

	subs = append(subs, results.FDNSA...)
	subs = append(subs, results.RDNS...)
	subs = append(subs, results.Results...)

	for _, subdomain := range subs {
		for _, value := range config.SubdomainsRegex.FindAllString(subdomain, -1) {
			subdomains <- sources.Subdomain{Source: source.Name(), Value: value}
		}
	}
}

func (source *Source) Name() string {
	return "bufferover"
}
