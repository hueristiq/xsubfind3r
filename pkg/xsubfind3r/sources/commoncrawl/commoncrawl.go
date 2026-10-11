package commoncrawl

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	hqgohttpheader "github.com/hueristiq/hq-lib-http-go/header"
	"github.com/spf13/cast"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type getIndexesResponse []struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	TimeGate string `json:"timegate"`
	CDXAPI   string `json:"cdx-api"`
	From     string `json:"from"`
	To       string `json:"to"`
}

type getPaginationResponse struct {
	Blocks   uint `json:"blocks"`
	PageSize uint `json:"pageSize"`
	Pages    uint `json:"pages"`
}

type getURLsResponse struct {
	URL   string `json:"url"`
	Error string `json:"error"`
}

type Source struct{}

func (s *Source) Name() (name string) {
	name = sources.COMMONCRAWL

	return
}

func (s *Source) Run(ctx context.Context, cfg *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		getIndexesReqURL := "https://index.commoncrawl.org/collinfo.json"

		getIndexesRes, err := cfg.Client.NewRequest().
			URL(getIndexesReqURL).
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

		var getIndexesResData getIndexesResponse

		if err = json.NewDecoder(getIndexesRes.Body).Decode(&getIndexesResData); err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  fmt.Errorf("failed to parse JSON response: %w", err),
			}

			results <- result

			sources.CloseBody(getIndexesRes.Body)

			return
		}

		sources.CloseBody(getIndexesRes.Body)

		year := time.Now().Year()
		years := make([]string, 0)
		maxYearsBack := 5

		for i := range maxYearsBack {
			years = append(years, strconv.Itoa(year-i))
		}

		searchIndexes := make(map[string]string)

		for _, year := range years {
			for _, ccIndex := range getIndexesResData {
				if strings.Contains(ccIndex.ID, year) {
					if _, ok := searchIndexes[year]; !ok {
						searchIndexes[year] = ccIndex.CDXAPI

						break
					}
				}
			}
		}

		for _, ccIndexAPI := range searchIndexes {
			getPaginationRes, err := cfg.Client.NewRequest().
				URL(ccIndexAPI).
				Param("url", "*."+domain+"/*").
				Param("output", "json").
				Param("fl", "url").
				Param("showNumPages", "true").
				SetHeader(hqgohttpheader.Host.String(), "index.commoncrawl.org").
				Do(ctx)
			if err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: s.Name(),
					Error:  fmt.Errorf("request failed: %w", err),
				}

				results <- result

				continue
			}

			var getPaginationResData getPaginationResponse

			if err = json.NewDecoder(getPaginationRes.Body).Decode(&getPaginationResData); err != nil {
				result := sources.Result{
					Type:   sources.ResultError,
					Source: s.Name(),
					Error:  fmt.Errorf("failed to parse JSON response: %w", err),
				}

				results <- result

				sources.CloseBody(getPaginationRes.Body)

				continue
			}

			sources.CloseBody(getPaginationRes.Body)

			if getPaginationResData.Pages < 1 {
				continue
			}

			for page := range getPaginationResData.Pages {
				getURLsRes, err := cfg.Client.NewRequest().
					URL(ccIndexAPI).
					Param("url", "*."+domain+"/*").
					Param("output", "json").
					Param("fl", "url").
					Param("page", cast.ToString(page)).
					SetHeader(hqgohttpheader.Host.String(), "index.commoncrawl.org").
					Do(ctx)
				if err != nil {
					result := sources.Result{
						Type:   sources.ResultError,
						Source: s.Name(),
						Error:  fmt.Errorf("request failed: %w", err),
					}

					results <- result

					continue
				}

				scanner := bufio.NewScanner(getURLsRes.Body)

				for scanner.Scan() {
					var getURLsResData getURLsResponse

					if err = json.Unmarshal(scanner.Bytes(), &getURLsResData); err != nil {
						result := sources.Result{
							Type:   sources.ResultError,
							Source: s.Name(),
							Error:  fmt.Errorf("failed to parse JSON response: %w", err),
						}

						results <- result

						continue
					}

					if getURLsResData.Error != "" {
						result := sources.Result{
							Type:   sources.ResultError,
							Source: s.Name(),
							Error:  fmt.Errorf("domain error: %s", getURLsResData.Error),
						}

						results <- result

						continue
					}

					subdomains := cfg.Extract(getURLsResData.URL)

					for _, subdomain := range subdomains {
						result := sources.Result{
							Type:   sources.ResultSubdomain,
							Source: s.Name(),
							Value:  subdomain,
						}

						results <- result
					}
				}

				if err = scanner.Err(); err != nil {
					result := sources.Result{
						Type:   sources.ResultError,
						Source: s.Name(),
						Error:  fmt.Errorf("failed to read response body: %w", err),
					}

					results <- result

					sources.CloseBody(getURLsRes.Body)

					continue
				}

				sources.CloseBody(getURLsRes.Body)
			}
		}
	}()

	return results
}

func (s *Source) NeedsKeys() (needs bool) {
	needs = false

	return
}

func (s *Source) UseKeys(keys ...string) {
}

var _ sources.Source = (*Source)(nil)

func New() (source sources.Source) {
	source = &Source{}

	return
}
