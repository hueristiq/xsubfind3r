package threatminer

import (
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/valyala/fasthttp"
)

type response struct {
	StatusCode    string   `json:"status_code"`
	StatusMessage string   `json:"status_message"`
	Results       []string `json:"results"`
}

type Source struct{}

func (source *Source) Run(config *sources.Configuration) (subdomains chan sources.Subdomain) {
	subdomains = make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		var (
			err error
			res *fasthttp.Response
		)

		res, err = httpclient.SimpleGet(fmt.Sprintf("https://api.threatminer.org/v2/domain.php?q=%s&rt=5", config.Domain))
		if err != nil {
			return
		}

		var results response

		if err = json.Unmarshal(res.Body(), &results); err != nil {
			return
		}

		for _, i := range results.Results {
			subdomains <- sources.Subdomain{Source: source.Name(), Value: i}
		}
	}()

	return
}

func (source *Source) Name() string {
	return "threatminer"
}
