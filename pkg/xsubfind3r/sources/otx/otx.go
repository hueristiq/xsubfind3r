package otx

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type getPassiveDNSResponse struct {
	Detail     string `json:"detail"`
	Error      string `json:"error"`
	PassiveDNS []struct {
		Hostname string `json:"hostname"`
	} `json:"passive_dns"`
}

type Source struct {
	keys sources.Keys
}

func (s *Source) Name() (name string) {
	name = sources.OpenThreatExchange

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

		key := ""

		if picked, err := s.keys.PickRandom(); err == nil {
			key = picked
		}

		getPassiveDNSReqURL := fmt.Sprintf("https://otx.alienvault.com/api/v1/indicators/domain/%s/passive_dns", domain)

		getPassiveDNSReq := cfg.Client.Request().
			URL(getPassiveDNSReqURL)

		if key != "" {
			getPassiveDNSReq = getPassiveDNSReq.SetHeader("X-OTX-API-KEY", key)
		}

		getPassiveDNSRes, err := getPassiveDNSReq.Context(ctx).Do()
		if err != nil {
			result := sources.ErrorResult(s.Name(), fmt.Errorf("request failed: %w", err))

			results <- result

			return
		}

		if err = sources.CheckResponse(getPassiveDNSRes.Std(), key); err != nil {
			result := sources.ErrorResult(s.Name(), err)

			results <- result

			_ = getPassiveDNSRes.Body.Close()

			return
		}

		var getPassiveDNSResData getPassiveDNSResponse

		if err = json.NewDecoder(getPassiveDNSRes.Body).Decode(&getPassiveDNSResData); err != nil {
			result := sources.ErrorResult(s.Name(), fmt.Errorf("failed to parse JSON response: %w", err))

			results <- result

			_ = getPassiveDNSRes.Body.Close()

			return
		}

		_ = getPassiveDNSRes.Body.Close()

		if getPassiveDNSResData.Error != "" {
			result := sources.ErrorResult(s.Name(), fmt.Errorf("domain error: %s, %s", getPassiveDNSResData.Detail, getPassiveDNSResData.Error))

			results <- result

			return
		}

		for _, record := range getPassiveDNSResData.PassiveDNS {
			subdomain, valid := cfg.Validator(record.Hostname)
			if !valid {
				continue
			}

			result := sources.SubdomainResult(s.Name(), subdomain)

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
