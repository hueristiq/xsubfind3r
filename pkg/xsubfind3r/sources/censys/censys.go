package censys

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/valyala/fasthttp"
)

type response struct {
	Results  []resultsq `json:"results"`
	Metadata struct {
		Pages int `json:"pages"`
	} `json:"metadata"`
}

type resultsq struct {
	Data  []string `json:"parsed.extensions.subject_alt_name.dns_names"`
	Data1 []string `json:"parsed.names"`
}

type Source struct{}

const maxCensysPages = 10

func (source *Source) Run(config *sources.Configuration) (subdomains chan sources.Subdomain) {
	subdomains = make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		var (
			key     string
			err     error
			res     *fasthttp.Response
			headers = map[string]string{
				"Content-Type": "application/json",
				"Accept":       "application/json",
			}
		)

		key, err = sources.PickRandom(config.Keys.Censys)
		if key == "" || err != nil {
			return
		}

		parts := strings.Split(key, ":")
		username := parts[0]
		password := parts[1]

		if username == "" || password == "" {
			return
		}

		currentPage := 1
		for {
			var reqData = []byte(`{"query":"` + config.Domain + `", "page":` + strconv.Itoa(currentPage) + `, "fields":["parsed.names","parsed.extensions.subject_alt_name.dns_names"], "flatten":true}`)

			reqURL := fmt.Sprintf("https://%s:%s@search.censys.io/api/v1/search/certificates", username, password)

			res, err = httpclient.Request(fasthttp.MethodPost, reqURL, "", headers, reqData)

			if err != nil {
				return
			}

			body := res.Body()

			var results response

			if err = json.Unmarshal(body, &results); err != nil {
				return
			}

			for _, i := range results.Results {
				for _, part := range i.Data {
					subdomains <- sources.Subdomain{Source: source.Name(), Value: part}
				}
				for _, part := range i.Data1 {
					subdomains <- sources.Subdomain{Source: source.Name(), Value: part}
				}
			}

			// Exit the censys enumeration if max pages is reached
			if currentPage >= results.Metadata.Pages || currentPage >= maxCensysPages {
				break
			}

			currentPage++
		}
	}()

	return
}

func (source *Source) Name() string {
	return "censys"
}
