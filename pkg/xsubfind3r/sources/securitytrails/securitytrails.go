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

				getSubdomainsReqCFG := &hqgohttp.RequestConfiguration{
					Headers: map[string]string{
						"Content-Type": "application/json",
						"APIKEY":       key,
					},
				}

				getSubdomainsRes, err = hqgohttp.Post(getSubdomainsReqURL, getSubdomainsReqBodyDataReader, getSubdomainsReqCFG) //nolint:bodyclose
			} else {
				getSubdomainsReqURL := fmt.Sprintf("https://api.securitytrails.com/v1/scroll/%s", scrollID)
				getSubdomainsReqCFG := &hqgohttp.RequestConfiguration{
					Headers: map[string]string{
						"Content-Type": "application/json",
						"APIKEY":       key,
					},
				}

				getSubdomainsRes, err = hqgohttp.Get(getSubdomainsReqURL, getSubdomainsReqCFG) //nolint:bodyclose
			}

			if err != nil && getSubdomainsRes.StatusCode == status.Forbidden.Int() {
				getSubdomainsReqURL := fmt.Sprintf("https://api.securitytrails.com/v1/domain/%s/subdomains", domain)
				getSubdomainsReqCFG := &hqgohttp.RequestConfiguration{
					Params: map[string]string{
						"children_only":    "false",
						"include_inactive": "true",
					},
					Headers: map[string]string{
						"Content-Type": "application/json",
						"APIKEY":       key,
					},
				}

				getSubdomainsRes, err = hqgohttp.Get(getSubdomainsReqURL, getSubdomainsReqCFG) //nolint:bodyclose
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
