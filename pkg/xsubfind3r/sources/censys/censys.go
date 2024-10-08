package censys

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/hueristiq/xsubfind3r/pkg/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type certSearchResponse struct {
	Code   int    `json:"code"`
	Status string `json:"status"`
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

type Source struct{}

func (source *Source) Run(config *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		key, err := config.Keys.Censys.PickRandom()
		if key == "" || err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		cursor := ""

		page := 1

		certSearchReqURL := "https://search.censys.io/api/v2/certificates/search"
		certSearchReqHeaders := map[string]string{
			"Authorization": "Basic " + base64.StdEncoding.EncodeToString([]byte(key)),
		}

		for {
			certSearchReqURL = fmt.Sprintf(certSearchReqURL+"?q=%s&per_page=%d", domain, maxPerPage)

			if cursor != "" {
				certSearchReqURL = certSearchReqURL + "&cursor=" + cursor
			}

			certSearchRes, err := httpclient.Get(certSearchReqURL, "", certSearchReqHeaders)
			if err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: source.Name(),
					Error:  err,
				}

				results <- result

				httpclient.DiscardResponse(certSearchRes)

				return
			}

			var certSearchResData certSearchResponse

			if err = json.NewDecoder(certSearchRes.Body).Decode(&certSearchResData); err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: source.Name(),
					Error:  err,
				}

				results <- result

				certSearchRes.Body.Close()

				return
			}

			certSearchRes.Body.Close()

			for _, hit := range certSearchResData.Result.Hits {
				for _, name := range hit.Names {
					result := sources.Result{
						Type:   sources.ResultSubdomain,
						Source: source.Name(),
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

func (source *Source) Name() string {
	return sources.CENSYS
}

const (
	maxCensysPages = 10
	maxPerPage     = 100
)
