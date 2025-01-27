package securitytrails

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	hqgohttp "go.source.hueristiq.com/http"
	"go.source.hueristiq.com/http/status"
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

func (source *Source) Run(cfg *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		key, err := cfg.Keys.SecurityTrails.PickRandom()
		if key == "" || err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var scrollID string

		for {
			var err error

			var getSubdomainsRes *http.Response

			if scrollID == "" {
				getSubdomainsReqURL := "https://api.securitytrails.com/v1/domains/list?include_ips=false&scroll=true"

				type getSubdomainsReqBody struct {
					Query string `json:"query"`
				}

				getSubdomainsReqBodyData := getSubdomainsReqBody{
					Query: fmt.Sprintf("apex_domain=%q", domain),
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

				getSubdomainsRes, err = hqgohttp.POST(getSubdomainsReqURL).AddHeader("Content-Type", "application/json").AddHeader("APIKEY", key).Body(getSubdomainsReqBodyDataReader).Send() //nolint:bodyclose
			} else {
				getSubdomainsReqURL := fmt.Sprintf("https://api.securitytrails.com/v1/scroll/%s", scrollID)

				getSubdomainsRes, err = hqgohttp.GET(getSubdomainsReqURL).AddHeader("Content-Type", "application/json").AddHeader("APIKEY", key).Send() //nolint:bodyclose
			}

			if err != nil && getSubdomainsRes.StatusCode == status.Forbidden.Int() {
				getSubdomainsReqURL := fmt.Sprintf("https://api.securitytrails.com/v1/domain/%s/subdomains?children_only=false&include_inactive=true", domain)

				getSubdomainsRes, err = hqgohttp.GET(getSubdomainsReqURL).AddHeader("Content-Type", "application/json").AddHeader("APIKEY", key).Send() //nolint:bodyclose
			}

			if err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: source.Name(),
					Error:  err,
				}

				results <- result

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

			scrollID = getSubdomainsResData.Meta.ScrollID

			if scrollID == "" {
				break
			}
		}
	}()

	return results
}

func (source *Source) Name() string {
	return sources.SECURITYTRAILS
}
