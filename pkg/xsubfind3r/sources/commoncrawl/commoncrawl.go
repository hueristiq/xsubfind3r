package commoncrawl

import (
	"bufio"
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"

	"github.com/hueristiq/xsubfind3r/pkg/httpclient"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/extractor"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type getIndexesResponse []struct {
	ID  string `json:"id"`
	API string `json:"cdx-API"`
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

func (source *Source) Run(_ *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		var err error

		var regex *regexp.Regexp

		regex, err = extractor.New(domain)
		if err != nil {
			result := sources.Result{
				Type:   sources.Error,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		getIndexesReqURL := "https://index.commoncrawl.org/collinfo.json"

		var getIndexesRes *http.Response

		getIndexesRes, err = httpclient.SimpleGet(getIndexesReqURL)
		if err != nil {
			result := sources.Result{
				Type:   sources.Error,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var getIndexesResData getIndexesResponse

		err = json.NewDecoder(getIndexesRes.Body).Decode(&getIndexesResData)
		if err != nil {
			result := sources.Result{
				Type:   sources.Error,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			getIndexesRes.Body.Close()

			return
		}

		getIndexesRes.Body.Close()

		for _, indexData := range getIndexesResData {
			getURLsReqHeaders := map[string]string{
				"Host": "index.commoncrawl.org",
			}

			getPaginationReqURL := fmt.Sprintf("%s?url=*.%s/*&output=json&fl=url&showNumPages=true", indexData.API, domain)

			var getPaginationRes *http.Response

			getPaginationRes, err = httpclient.SimpleGet(getPaginationReqURL)
			if err != nil {
				result := sources.Result{
					Type:   sources.Error,
					Source: source.Name(),
					Error:  err,
				}

				results <- result

				continue
			}

			var getPaginationData getPaginationResponse

			err = json.NewDecoder(getPaginationRes.Body).Decode(&getPaginationData)
			if err != nil {
				result := sources.Result{
					Type:   sources.Error,
					Source: source.Name(),
					Error:  err,
				}

				results <- result

				getPaginationRes.Body.Close()

				continue
			}

			getPaginationRes.Body.Close()

			if getPaginationData.Pages < 1 {
				continue
			}

			for page := uint(0); page < getPaginationData.Pages; page++ {
				getURLsReqURL := fmt.Sprintf("%s?url=*.%s/*&output=json&fl=url&page=%d", indexData.API, domain, page)

				var getURLsRes *http.Response

				getURLsRes, err = httpclient.Get(getURLsReqURL, "", getURLsReqHeaders)
				if err != nil {
					result := sources.Result{
						Type:   sources.Error,
						Source: source.Name(),
						Error:  err,
					}

					results <- result

					continue
				}

				scanner := bufio.NewScanner(getURLsRes.Body)

				for scanner.Scan() {
					line := scanner.Text()
					if line == "" {
						continue
					}

					var getURLsResData getURLsResponse

					err = json.Unmarshal(scanner.Bytes(), &getURLsResData)
					if err != nil {
						result := sources.Result{
							Type:   sources.Error,
							Source: source.Name(),
							Error:  err,
						}

						results <- result

						continue
					}

					if getURLsResData.Error != "" {
						result := sources.Result{
							Type:   sources.Error,
							Source: source.Name(),
							Error:  fmt.Errorf("%s", getURLsResData.Error),
						}

						results <- result

						continue
					}

					match := regex.FindAllString(getURLsResData.URL, -1)

					for _, subdomain := range match {
						result := sources.Result{
							Type:   sources.Subdomain,
							Source: source.Name(),
							Value:  subdomain,
						}

						results <- result
					}
				}

				if err = scanner.Err(); err != nil {
					result := sources.Result{
						Type:   sources.Error,
						Source: source.Name(),
						Error:  err,
					}

					results <- result

					getURLsRes.Body.Close()

					continue
				}

				getURLsRes.Body.Close()
			}
		}
	}()

	return results
}

func (source *Source) Name() string {
	return "commoncrawl"
}
