package github

import (
	"context"
	"fmt"
	"net/http"

	hqgosource "github.com/hueristiq/xsubfind3r/internal/source"
	"github.com/hueristiq/xsubfind3r/internal/source/github"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type Source struct {
	keys sources.Keys
}

func (s *Source) Name() string {
	return sources.Github
}

func (s *Source) NeedsKeys() bool {
	return true
}

func (s *Source) UseKeys(keys ...string) {
	s.keys = append(s.keys, keys...)
}

func (s *Source) Run(ctx context.Context, cfg *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		emit := func(result sources.Result) bool {
			select {
			case results <- result:
				return true
			case <-ctx.Done():
				return false
			}
		}

		client, err := github.New(hqgosource.DoerFunc(func(request *http.Request) (*http.Response, error) {
			return cfg.Client.Do(request)
		}), s.keys)
		if err != nil {
			emit(sources.ErrorResult(s.Name(), err))

			return
		}

		for result := range client.Search(ctx, github.Query{Term: fmt.Sprintf("%q", domain)}) {
			if result.Err != nil {
				if !emit(sources.ErrorResult(s.Name(), result.Err)) {
					return
				}

				continue
			}

			if result.Warning != "" {
				if !emit(sources.WarningResult(s.Name(), result.Warning)) {
					return
				}

				continue
			}

			for _, candidate := range sources.ExtractSubdomains(cfg.Extractor, result.Value) {
				subdomain, valid := cfg.Validator(candidate)
				if !valid {
					continue
				}

				if !emit(sources.SubdomainResult(s.Name(), subdomain)) {
					return
				}
			}
		}
	}()

	return results
}

var _ sources.Source = (*Source)(nil)

func New() sources.Source {
	return &Source{keys: make(sources.Keys, 0)}
}
