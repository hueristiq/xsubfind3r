package shodan

import (
	"encoding/json"
	"fmt"

	hqgohttp "github.com/hueristiq/hq-go-http"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type getDNSResponse struct {
	Domain     string   `json:"domain"`
	Subdomains []string `json:"subdomains"`
	Result     int      `json:"result"`
	Error      string   `json:"error"`
}

type Source struct {
	keys sources.Keys
}

func (s *Source) Name() (name string) {
	name = sources.SHODAN

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

		getDNSReqURL := "https://api.shodan.io/dns/domain/" + domain
		getDNSReqCFG := &hqgohttp.RequestConfiguration{
			Params: map[string]string{
				"key": key,
			},
		}

		getDNSRes, err := hqgohttp.Get(getDNSReqURL, getDNSReqCFG)
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  fmt.Errorf("request failed: %w", err),
			}

			results <- result

			return
		}

		var getDNSResData getDNSResponse

		if err = json.NewDecoder(getDNSRes.Body).Decode(&getDNSResData); err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  fmt.Errorf("failed to parse JSON response: %w", err),
			}

			results <- result

			getDNSRes.Body.Close()

			return
		}

		getDNSRes.Body.Close()

		for _, subdomain := range getDNSResData.Subdomains {
			result := sources.Result{
				Type:   sources.ResultSubdomain,
				Source: s.Name(),
				Value:  fmt.Sprintf("%s.%s", subdomain, domain),
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
