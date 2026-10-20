package fofa

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type getSubdomainsResponse struct {
	Error   bool     `json:"error"`
	ErrMsg  string   `json:"errmsg"`
	Size    int      `json:"size"`
	Results []string `json:"results"`
}

type Source struct {
	keys sources.Keys
}

func (s *Source) Name() (name string) {
	name = sources.Fofa

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

		if len(s.keys) < 2 {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  errors.New("no keys configured"),
			}

			results <- result

			return
		}

		email := s.keys[0]
		key := s.keys[1]

		qbase64 := base64.StdEncoding.EncodeToString([]byte(fmt.Sprintf("domain=%q", domain)))

		getSubdomainsReqURL := "https://fofa.info/api/v1/search/all"

		getSubdomainsRes, err := cfg.Client.NewRequest().
			URL(getSubdomainsReqURL).
			Param("full", "true").
			Param("fields", "host").
			Param("page", "1").
			Param("size", "10000").
			Param("email", email).
			Param("key", key).
			Param("qbase64", qbase64).
			Do(ctx)
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  fmt.Errorf("request failed: %w", sources.RedactError(err, email, key)),
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

		if getSubdomainsResData.Error {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  fmt.Errorf("%s", getSubdomainsResData.ErrMsg),
			}

			results <- result

			return
		}

		for _, host := range getSubdomainsResData.Results {
			subdomains := cfg.Extract(host)

			for _, subdomain := range subdomains {
				result := sources.Result{
					Type:   sources.ResultSubdomain,
					Source: s.Name(),
					Value:  subdomain,
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
