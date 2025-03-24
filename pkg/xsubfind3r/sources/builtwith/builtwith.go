package builtwith

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	hqgohttp "go.source.hueristiq.com/http"
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

type Source struct{}

func (source *Source) Run(cfg *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		key, err := cfg.Keys.BuiltWith.PickRandom()
		if key == "" || err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
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
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var getDomainInfoResData getDomainInfoResponse

		if err = json.NewDecoder(getDomainInfoRes.Body).Decode(&getDomainInfoResData); err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
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
					Source: source.Name(),
					Error:  fmt.Errorf("%w: %s", errStatic, entry.Message),
				}

				results <- result
			}

			return
		}

		for _, item := range getDomainInfoResData.Results {
			for _, path := range item.Result.Paths {
				result := sources.Result{
					Type:   sources.ResultSubdomain,
					Source: source.Name(),
					Value:  path.SubDomain + "." + path.Domain,
				}

				results <- result
			}
		}
	}()

	return results
}

func (source *Source) Name() string {
	return sources.BUILTWITH
}

var errStatic = errors.New("something went wrong")
