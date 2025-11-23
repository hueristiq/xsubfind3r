package censys

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	hqgohttp "github.com/hueristiq/hq-go-http"
	hqgohttpheader "github.com/hueristiq/hq-go-http/header"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/spf13/cast"
)

type certSearchResponse struct {
	Code   int    `json:"code"`
	Status string `json:"status"`
	Error  string `json:"error"`
	Result struct {
		Query      string  `json:"query"`
		Total      float64 `json:"total"`
		DurationMS int     `json:"duration_ms"`
		Hits       []struct {
			Parsed struct {
				ValidityPeriod struct {
					NotAfter  string `json:"not_after"`
					NotBefore string `json:"not_before"`
				} `json:"validity_period"`
				SubjectDN string `json:"subject_dn"`
				IssuerDN  string `json:"issuer_dn"`
			} `json:"parsed"`
			Names             []string `json:"names"`
			FingerprintSha256 string   `json:"fingerprint_sha256"`
		} `json:"hits"`
		Links struct {
			Next string `json:"next"`
			Prev string `json:"prev"`
		} `json:"links"`
	} `json:"result"`
}

type Source struct {
	keys sources.Keys
}

func (s *Source) Name() (name string) {
	name = sources.CENSYS

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

		page := 1
		cursor := ""

		certSearchReqURL := "https://search.censys.io/api/v2/certificates/search"

		for {
			certSearchReqCFG := &hqgohttp.RequestConfiguration{
				Params: map[string]string{
					"q":        domain,
					"per_page": cast.ToString(maxPerPage),
				},
				Headers: []hqgohttp.Header{
					hqgohttp.NewSetHeader(hqgohttpheader.Authorization.String(), "Basic "+base64.StdEncoding.EncodeToString([]byte(key))),
				},
			}

			if cursor != "" {
				certSearchReqCFG.Params["cursor"] = cursor
			}

			certSearchRes, err := hqgohttp.Get(certSearchReqURL, certSearchReqCFG)
			if err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: s.Name(),
					Error:  fmt.Errorf("request failed: %w", err),
				}

				results <- result

				return
			}

			var certSearchResData certSearchResponse

			if err = json.NewDecoder(certSearchRes.Body).Decode(&certSearchResData); err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: s.Name(),
					Error:  fmt.Errorf("failed to parse JSON response: %w", err),
				}

				results <- result

				certSearchRes.Body.Close()

				return
			}

			certSearchRes.Body.Close()

			if certSearchResData.Error != "" {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: s.Name(),
					Error:  fmt.Errorf("domain error: %s, %s", certSearchResData.Status, certSearchResData.Error),
				}

				results <- result

				return
			}

			for _, hit := range certSearchResData.Result.Hits {
				for _, name := range hit.Names {
					result := sources.Result{
						Type:   sources.ResultSubdomain,
						Source: s.Name(),
						Value:  name,
					}

					results <- result
				}
			}

			cursor = certSearchResData.Result.Links.Next

			if cursor == "" || page >= maxCensysPages {
				break
			}

			page++
		}
	}()

	return results
}

func (s *Source) UseKeys(keys ...string) {
	s.keys = append(s.keys, keys...)
}

const (
	maxCensysPages = 10
	maxPerPage     = 100
)

var _ sources.Source = (*Source)(nil)

func New() (source sources.Source) {
	source = &Source{
		keys: make(sources.Keys, 0),
	}

	return
}
