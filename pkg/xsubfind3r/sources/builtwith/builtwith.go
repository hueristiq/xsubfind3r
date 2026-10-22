package builtwith

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type getDomainInfoResponse struct {
	Results []struct {
		Result struct {
			Paths []struct {
				Domain    string `json:"Domain"`
				URL       string `json:"Url"`
				SubDomain string `json:"SubDomain"`
			} `json:"Paths"`
		} `json:"Result"`
	} `json:"Results"`
	Errors []struct {
		Lookup  string `json:"Lookup"`
		Message string `json:"Message"`
	} `json:"Errors"`
}

type Source struct {
	keys sources.Keys
}

func (s *Source) Name() (name string) {
	name = sources.Builtwith

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
			result := sources.ErrorResult(s.Name(), fmt.Errorf("failed to select key: %w", err))

			results <- result

			return
		}

		getDomainInfoReqURL := "https://api.builtwith.com/v21/api.json"

		getDomainInfoRes, err := cfg.Client.NewRequest().
			URL(getDomainInfoReqURL).
			Param("KEY", key).
			Param("HIDETEXT", "yes").
			Param("HIDEDL", "yes").
			Param("NOLIVE", "yes").
			Param("NOMETA", "yes").
			Param("NOPII", "yes").
			Param("NOATTR", "yes").
			Param("LOOKUP", domain).
			Do(ctx)
		if err != nil {
			result := sources.ErrorResult(s.Name(), fmt.Errorf("request failed: %w", sources.RedactError(err, key)))

			results <- result

			return
		}

		if err = sources.CheckResponse(getDomainInfoRes, key); err != nil {
			result := sources.ErrorResult(s.Name(), err)

			results <- result

			_ = getDomainInfoRes.Body.Close()

			return
		}

		var getDomainInfoResData getDomainInfoResponse

		if err = json.NewDecoder(getDomainInfoRes.Body).Decode(&getDomainInfoResData); err != nil {
			result := sources.ErrorResult(s.Name(), fmt.Errorf("failed to parse JSON response: %w", err))

			results <- result

			_ = getDomainInfoRes.Body.Close()

			return
		}

		_ = getDomainInfoRes.Body.Close()

		if len(getDomainInfoResData.Errors) > 0 {
			for _, entry := range getDomainInfoResData.Errors {
				result := sources.ErrorResult(s.Name(), fmt.Errorf("domain error: %s", entry.Message))

				results <- result
			}

			return
		}

		for _, item := range getDomainInfoResData.Results {
			for _, path := range item.Result.Paths {
				subdomain, valid := cfg.Validator(path.SubDomain + "." + path.Domain)
				if !valid {
					continue
				}

				result := sources.SubdomainResult(s.Name(), subdomain)

				results <- result
			}
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
