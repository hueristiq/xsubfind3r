package fullhunt

import (
	"encoding/json"
	"fmt"

	hqgohttp "github.com/hueristiq/hq-go-http"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type getSubdomainsResponse struct {
	Hosts   []string `json:"hosts"`
	Message string   `json:"message"`
	Status  int      `json:"status"`
}

type Source struct {
	keys sources.Keys
}

func (s *Source) Name() (name string) {
	name = sources.FULLHUNT

	return
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

		getSubdomainsReqURL := fmt.Sprintf("https://fullhunt.io/api/v1/domain/%s/subdomains", domain)
		getSubdomainsReqCFG := &hqgohttp.RequestConfiguration{
			Headers: []hqgohttp.Header{
				hqgohttp.NewSetHeader("X-API-KEY", key),
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

		err = json.NewDecoder(getSubdomainsRes.Body).Decode(&getSubdomainsResData)
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  fmt.Errorf("failed to parse JSON response: %w", err),
			}

			results <- result

			getSubdomainsRes.Body.Close()

			return
		}

		getSubdomainsRes.Body.Close()

		for _, subdomain := range getSubdomainsResData.Hosts {
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

func (s *Source) UseKeys(keys ...string) {
	s.keys = append(s.keys, keys...)
}

var _ sources.Source = (*Source)(nil)

func New() (source sources.Source) {
	source = &Source{
		keys: make(sources.Keys, 0),
	}

	return
}
