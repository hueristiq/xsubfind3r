package securitytrails

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hueristiq/xsubfind3r/pkg/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type getSubdomainsResponse struct {
	Meta struct {
		ScrollID string `json:"scroll_id"`
	} `json:"meta"`
	Records []struct {
		Hostname string `json:"hostname"`
	} `json:"records"`
	Subdomains []string `json:"subdomains"`
}

type Source struct{}

func (source *Source) Run(config *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		key, err := config.Keys.SecurityTrails.PickRandom()
		if key == "" || err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var scrollId string

		getSubdomainsReqHeaders := map[string]string{
			"Content-Type": "application/json",
			"APIKEY":       key,
		}

		for {
			var err error

			var getSubdomainsRes *http.Response

			if scrollId == "" {
				getSubdomainsReqURL := "https://api.securitytrails.com/v1/domains/list?include_ips=false&scroll=true"

				type getSubdomainsReqBody struct {
					Query string `json:"query"`
				}

				getSubdomainsReqBodyData := getSubdomainsReqBody{
					Query: fmt.Sprintf("apex_domain='%s'", domain),
				}

				var getSubdomainsReqBodyDataBytes []byte

				getSubdomainsReqBodyDataBytes, err = json.Marshal(getSubdomainsReqBodyData)
				if err != nil {
					result := sources.Result{
						Type:   sources.ResultError,
						Source: source.Name(),
						Error:  err,
					}

					results <- result

					return
				}

				getSubdomainsReqBodyDataReader := bytes.NewReader(getSubdomainsReqBodyDataBytes)

				getSubdomainsRes, err = httpclient.Post(getSubdomainsReqURL, "", getSubdomainsReqHeaders, getSubdomainsReqBodyDataReader)
			} else {
				getSubdomainsReqURL := fmt.Sprintf("https://api.securitytrails.com/v1/scroll/%s", scrollId)

				getSubdomainsRes, err = httpclient.Get(getSubdomainsReqURL, "", getSubdomainsReqHeaders)
			}

			if err != nil && getSubdomainsRes.StatusCode == 403 {
				getSubdomainsReqURL := fmt.Sprintf("https://api.securitytrails.com/v1/domain/%s/subdomains?children_only=false&include_inactive=true", domain)

				getSubdomainsRes, err = httpclient.Get(getSubdomainsReqURL, "", getSubdomainsReqHeaders)
			}

			if err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: source.Name(),
					Error:  err,
				}

				results <- result

				httpclient.DiscardResponse(getSubdomainsRes)

				return
			}

			var getSubdomainsResData getSubdomainsResponse

			if err = json.NewDecoder(getSubdomainsRes.Body).Decode(&getSubdomainsResData); err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: source.Name(),
					Error:  err,
				}

				results <- result

				getSubdomainsRes.Body.Close()

				return
			}

			getSubdomainsRes.Body.Close()

			for _, record := range getSubdomainsResData.Records {
				result := sources.Result{
					Type:   sources.ResultSubdomain,
					Source: source.Name(),
					Value:  record.Hostname,
				}

				results <- result
			}

			for _, subdomain := range getSubdomainsResData.Subdomains {
				if strings.HasSuffix(subdomain, ".") {
					subdomain += domain
				} else {
					subdomain = subdomain + "." + domain
				}

				result := sources.Result{
					Type:   sources.ResultSubdomain,
					Source: source.Name(),
					Value:  subdomain,
				}

				results <- result
			}

			scrollId = getSubdomainsResData.Meta.ScrollID

			if scrollId == "" {
				break
			}
		}
	}()

	return results
}

func (source *Source) Name() string {
	return sources.SECURITYTRAILS
}
