package builtwith

import (
	"encoding/json"
	"fmt"

	hqgohttp "github.com/hueristiq/hq-go-http"
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
	name = sources.BUILTWITH

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

		getDomainInfoReqURL := "https://api.builtwith.com/v21/api.json"
		getDomainInfoReqCFG := &hqgohttp.RequestConfiguration{
			Params: map[string]string{
				"KEY":      key,
				"HIDETEXT": "yes",
				"HIDEDL":   "yes",
				"NOLIVE":   "yes",
				"NOMETA":   "yes",
				"NOPII":    "yes",
				"NOATTR":   "yes",
				"LOOKUP":   domain,
			},
		}

		getDomainInfoRes, err := hqgohttp.Get(getDomainInfoReqURL, getDomainInfoReqCFG)
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  fmt.Errorf("request failed: %w", err),
			}

			results <- result

			return
		}

		var getDomainInfoResData getDomainInfoResponse

		if err = json.NewDecoder(getDomainInfoRes.Body).Decode(&getDomainInfoResData); err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  fmt.Errorf("failed to parse JSON response: %w", err),
			}

			results <- result

			getDomainInfoRes.Body.Close()

			return
		}

		getDomainInfoRes.Body.Close()

		if len(getDomainInfoResData.Errors) > 0 {
			for _, entry := range getDomainInfoResData.Errors {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: s.Name(),
					Error:  fmt.Errorf("domain error: %s", entry.Message),
				}

				results <- result
			}

			return
		}

		for _, item := range getDomainInfoResData.Results {
			for _, path := range item.Result.Paths {
				result := sources.Result{
					Type:   sources.ResultSubdomain,
					Source: s.Name(),
					Value:  path.SubDomain + "." + path.Domain,
				}

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
