package github

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"strings"
	"sync"
	"time"

	hqgohttpheader "github.com/hueristiq/hq-lib-http-go/header"
	hqgohttpheaderutils "github.com/hueristiq/hq-lib-http-go/header/utils"
	hqgohttpstatus "github.com/hueristiq/hq-lib-http-go/status"
	"github.com/spf13/cast"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

// searchAPIURL is the GitHub code-search API endpoint.
const searchAPIURL = "https://api.github.com/search/code"

type codeSearchResponse struct {
	TotalCount int `json:"total_count"`
	Items      []struct {
		Name        string `json:"name"`
		HTMLURL     string `json:"html_url"`
		TextMatches []struct {
			Fragment string `json:"fragment"`
		} `json:"text_matches"`
	} `json:"items"`
}

// Source enumerates subdomains through the GitHub code-search API.
type Source struct {
	keys sources.Keys
}

func (s *Source) Name() (name string) {
	name = sources.GITHUB

	return
}

func (s *Source) Run(ctx context.Context, cfg *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		if len(s.keys) == 0 {
			return
		}

		params := url.Values{}
		params.Set("per_page", "100")
		params.Set("q", fmt.Sprintf("%q", domain))
		params.Set("sort", "created")
		params.Set("order", "asc")

		searchReqURL := searchAPIURL + "?" + params.Encode()

		s.Enumerate(ctx, searchReqURL, cfg, results, NewKeyManager(s.keys))
	}()

	return results
}

func (s *Source) Enumerate(ctx context.Context, searchReqURL string, cfg *sources.Configuration, results chan sources.Result, keysManager *KeysManager) {
	token := keysManager.GetCurrentKey()

	if token.RetryAfter > 0 {
		remaining := time.Duration(token.RetryAfter)*time.Second - time.Since(token.ExceededTime)

		if remaining > 0 {
			if !waitSleep(ctx, remaining) {
				return
			}
		}

		s.Enumerate(ctx, searchReqURL, cfg, results, keysManager)

		return
	}

	codeSearchRes, err := cfg.Client.NewRequest().
		URL(searchReqURL).
		SetHeader(hqgohttpheader.Accept.String(), "application/vnd.github.v3.text-match+json").
		SetHeader(hqgohttpheader.Authorization.String(), "token "+token.Value).
		Do(ctx)

	isForbidden := codeSearchRes != nil && codeSearchRes.StatusCode == hqgohttpstatus.Forbidden.Int()

	if err != nil && !isForbidden {
		result := sources.Result{
			Type:   sources.ResultError,
			Source: s.Name(),
			Error:  fmt.Errorf("request failed: %w", err),
		}

		results <- result

		return
	}

	ratelimitRemaining := cast.ToInt64(
		codeSearchRes.Header.Get(hqgohttpheader.XRateLimitRemaining.String()),
	)

	if isForbidden && ratelimitRemaining == 0 {
		retryAfterSeconds := cast.ToInt64(codeSearchRes.Header.Get(hqgohttpheader.RetryAfter.String()))

		if retryAfterSeconds <= 0 {
			retryAfterSeconds = rateLimitResetRemaining(codeSearchRes.Header.Get("X-Ratelimit-Reset"))
		}

		if retryAfterSeconds <= 0 {
			sources.CloseBody(codeSearchRes.Body)

			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  errors.New("rate limit reached with no recovery hint; try again later"),
			}

			results <- result

			return
		}

		keysManager.SetCurrentKeyExceeded(retryAfterSeconds)

		sources.CloseBody(codeSearchRes.Body)

		s.Enumerate(ctx, searchReqURL, cfg, results, keysManager)

		return
	}

	var codeSearchResData codeSearchResponse

	if err = json.NewDecoder(codeSearchRes.Body).Decode(&codeSearchResData); err != nil {
		result := sources.Result{
			Type:   sources.ResultError,
			Source: s.Name(),
			Error:  fmt.Errorf("failed to parse JSON response: %w", err),
		}

		results <- result

		sources.CloseBody(codeSearchRes.Body)

		return
	}

	sources.CloseBody(codeSearchRes.Body)

	for _, item := range codeSearchResData.Items {
		getRawContentReqURL := strings.ReplaceAll(item.HTMLURL, "https://github.com/", "https://raw.githubusercontent.com/")
		getRawContentReqURL = strings.ReplaceAll(getRawContentReqURL, "/blob/", "/")

		getRawContentRes, err := cfg.Client.NewRequest().
			URL(getRawContentReqURL).
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

		if getRawContentRes.StatusCode != hqgohttpstatus.OK.Int() {
			sources.CloseBody(getRawContentRes.Body)

			continue
		}

		scanner := bufio.NewScanner(getRawContentRes.Body)

		for scanner.Scan() {
			line := scanner.Text()

			if line == "" {
				continue
			}

			for _, subdomain := range cfg.Extract(line) {
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

			sources.CloseBody(getRawContentRes.Body)

			return
		}

		sources.CloseBody(getRawContentRes.Body)

		for _, match := range item.TextMatches {
			for _, subdomain := range cfg.Extract(match.Fragment) {
				result := sources.Result{
					Type:   sources.ResultSubdomain,
					Source: s.Name(),
					Value:  subdomain,
				}

				results <- result
			}
		}
	}

	links := hqgohttpheaderutils.Parse(codeSearchRes.Header.Get(hqgohttpheader.Link.String()))

	for _, link := range links {
		if link.Rel != "next" {
			continue
		}

		nextURL, err := url.QueryUnescape(link.URL)
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		s.Enumerate(ctx, nextURL, cfg, results, keysManager)
	}
}

func (s *Source) NeedsKeys() (needs bool) {
	needs = true

	return
}

func (s *Source) UseKeys(keys ...string) {
	s.keys = append(s.keys, keys...)
}

// ManagedKey is a single API key tracked by the [KeysManager].
type ManagedKey struct {
	ExceededTime time.Time
	RetryAfter   int64
	Value        string
}

// KeysManager rotates GitHub API keys in round-robin order, tracking which
// keys are rate-limited until when.
type KeysManager struct {
	mu      sync.Mutex
	current int
	last    int
	pool    []ManagedKey
}

// NewKeyManager returns a manager rotating through keys.
func NewKeyManager(keys []string) (manager *KeysManager) {
	pool := make([]ManagedKey, len(keys))

	for i, key := range keys {
		pool[i] = ManagedKey{Value: key}
	}

	manager = &KeysManager{
		pool: pool,
	}

	return
}

// GetCurrentKey returns the next key in the rotation, resetting any keys
// whose rate-limit cooldown has elapsed.
func (m *KeysManager) GetCurrentKey() (key *ManagedKey) {
	m.mu.Lock()
	defer m.mu.Unlock()

	for i := range m.pool {
		key := &m.pool[i]

		if key.RetryAfter > 0 && time.Since(key.ExceededTime) > time.Duration(key.RetryAfter)*time.Second {
			key.ExceededTime = time.Time{}
			key.RetryAfter = 0
		}
	}

	if m.current >= len(m.pool) {
		m.current %= len(m.pool)
	}

	m.last = m.current

	key = &m.pool[m.current]

	m.current++

	return
}

// SetCurrentKeyExceeded marks the key most recently handed out by
// [GetCurrentKey] as rate-limited for retryAfter seconds.
func (m *KeysManager) SetCurrentKeyExceeded(retryAfter int64) {
	m.mu.Lock()
	defer m.mu.Unlock()

	key := &m.pool[m.last]

	if key.RetryAfter == 0 {
		key.ExceededTime = time.Now()
		key.RetryAfter = retryAfter
	}
}

// rateLimitResetRemaining parses an X-RateLimit-Reset header value into the
// seconds remaining until the rate-limit window resets, or 0 when the value
// carries no future timestamp.
func rateLimitResetRemaining(raw string) (remaining int64) {
	reset := cast.ToInt64(raw)

	if reset > 0 {
		remaining = reset - time.Now().Unix()
	}

	return
}

// waitSleep sleeps for d unless ctx ends first; it reports whether the
// sleep completed.
func waitSleep(ctx context.Context, d time.Duration) (completed bool) {
	timer := time.NewTimer(d)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}

var _ sources.Source = (*Source)(nil)

func New() (source sources.Source) {
	source = &Source{
		keys: make(sources.Keys, 0),
	}

	return
}
