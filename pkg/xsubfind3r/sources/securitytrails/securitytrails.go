package securitytrails

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	hqgohttpheader "github.com/hueristiq/hq-lib-http-go/header"
	hqgohttpmime "github.com/hueristiq/hq-lib-http-go/mime"

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
	SubdomainCount int      `json:"subdomain_count"`
	Subdomains     []string `json:"subdomains"`
}

type Source struct {
	keys sources.Keys
}

func (s *Source) Name() (name string) {
	name = sources.Securitytrails

	return
}

func (s *Source) NeedsKeys() (needs bool) {
	needs = true

	return
}

func (s *Source) UseKeys(keys ...string) {
	s.keys = append(s.keys, keys...)
}

func (s *Source) Run(ctx context.Context, cfg *sources.Configuration, domain string) <-chan sources.Result {
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

		getSubdomainsRes, err := cfg.Client.NewRequest().
			URL(getSubdomainsReqURL).
			Param("children_only", "false").
			Param("include_inactive", "true").
			SetHeader(hqgohttpheader.Accept.String(), hqgohttpmime.JSON.String()).
			SetHeader("APIKEY", key).
			Do(ctx)
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

			_ = getSubdomainsRes.Body.Close()

			return
		}

		_ = getSubdomainsRes.Body.Close()

		for _, subdomain := range getSubdomainsResData.Subdomains {
			if strings.HasSuffix(subdomain, ".") {
				subdomain += domain
			} else {
				subdomain = subdomain + "." + domain
			}

			subdomain, valid := cfg.Validator(subdomain)
			if !valid {
				continue
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
