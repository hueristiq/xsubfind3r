package arquivo

import (
	"context"
	"net/http"

	hqgosource "github.com/hueristiq/xsubfind3r/internal/source"
	"github.com/hueristiq/xsubfind3r/internal/source/arquivo"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

type Source struct{}

func (s *Source) Name() string {
	return sources.Arquivo
}

func (s *Source) NeedsKeys() bool {
	return false
}

func (s *Source) UseKeys(...string) {}

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

		client, err := arquivo.New(hqgosource.DoerFunc(func(request *http.Request) (*http.Response, error) {
			return cfg.Client.Do(request)
		}))
		if err != nil {
			emit(sources.ErrorResult(s.Name(), err))

			return
		}

		for result := range client.Search(ctx, domain) {
			if result.Err != nil {
				if !emit(sources.ErrorResult(s.Name(), result.Err)) {
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
	return &Source{}
}
