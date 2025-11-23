package securitytrails

import (
	"encoding/json"
	"fmt"
	"strings"

	hqgohttp "github.com/hueristiq/hq-go-http"
	hqgohttpheader "github.com/hueristiq/hq-go-http/header"
	hqgohttpmime "github.com/hueristiq/hq-go-http/mime"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type getSubdomainsResponse struct {
	Endpoint string `json:"endpoint"`
	Meta     struct {
		ScrollID     string `json:"scroll_id"`
		LimitReached string `json:"limit_reached"`
	} `json:"meta"`
	Records []struct {
		Hostname string `json:"hostname"`
	} `json:"records"`
	SubdomainCount bool     `json:"subdomain_count"`
	Subdomains     []string `json:"subdomains"`
}

type Source struct {
	keys sources.Keys
}

func (s *Source) Name() (name string) {
	name = sources.SECURITYTRAILS

	return
}

func (s *Source) UseKeys(keys ...string) {
	s.keys = append(s.keys, keys...)
}

func (s *Source) Run(cfg *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		key, err := s.keys.PickRandom()
		if key == "" || err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  fmt.Errorf("failed to select key: %w", err),
			}

			results <- result

			return
		}

		getSubdomainsReqURL := fmt.Sprintf("https://api.securitytrails.com/v1/domain/%s/subdomains", domain)
		getSubdomainsReqCFG := &hqgohttp.RequestConfiguration{
			Params: map[string]string{
				"children_only":    "false",
				"include_inactive": "true",
			},
			Headers: []hqgohttp.Header{
				hqgohttp.NewSetHeader(hqgohttpheader.Accept.String(), hqgohttpmime.JSON.String()),
				hqgohttp.NewSetHeader("APIKEY", key),
			},
		}

		getSubdomainsRes, err := hqgohttp.Get(getSubdomainsReqURL, getSubdomainsReqCFG)
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  fmt.Errorf("request failed: %w", err),
			}

			results <- result

			return
		}

		var getSubdomainsResData getSubdomainsResponse

		if err = json.NewDecoder(getSubdomainsRes.Body).Decode(&getSubdomainsResData); err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  fmt.Errorf("failed to parse JSON response: %w", err),
			}

			results <- result

			getSubdomainsRes.Body.Close()
		}

		getSubdomainsRes.Body.Close()

		for _, subdomain := range getSubdomainsResData.Subdomains {
			if strings.HasSuffix(subdomain, ".") {
				subdomain += domain
			} else {
				subdomain = subdomain + "." + domain
			}

			result := sources.Result{
				Type:   sources.ResultSubdomain,
				Source: s.Name(),
				Value:  subdomain,
			}

			results <- result
		}
	}()

	return results
}

var _ sources.Source = (*Source)(nil)

func New() (source sources.Source) {
	source = &Source{
		keys: make(sources.Keys, 0),
	}

	return
}
