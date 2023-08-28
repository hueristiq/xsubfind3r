package xsubfind3r

import (
	"strings"
	"sync"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/anubis"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/bevigil"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/chaos"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/commoncrawl"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/crtsh"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/fullhunt"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/github"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/hackertarget"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/intelx"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/otx"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/shodan"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/urlscan"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/wayback"
)

type Finder struct {
	Sources              map[string]sources.Source
	SourcesConfiguration *sources.Configuration
}

func (finder *Finder) Find(domain string) (results chan sources.Result) {
	results = make(chan sources.Result)

	go func() {
		defer close(results)

		seenSubdomains := &sync.Map{}

		wg := &sync.WaitGroup{}

		for _, source := range finder.Sources {
			wg.Add(1)

			go func(source sources.Source) {
				defer wg.Done()

				sResults := source.Run(finder.SourcesConfiguration, domain)

				for sResult := range sResults {
					if sResult.Type == sources.Subdomain {
						sResult.Value = strings.ToLower(sResult.Value)
						sResult.Value = strings.ReplaceAll(sResult.Value, "*.", "")

						_, loaded := seenSubdomains.LoadOrStore(sResult.Value, struct{}{})
						if loaded {
							continue
						}
					}

					results <- sResult
				}
			}(source)
		}

		wg.Wait()
	}()

	return
}

func New(options *Options) (finder *Finder) {
	finder = &Finder{
		Sources: map[string]sources.Source{},
		SourcesConfiguration: &sources.Configuration{
			Keys: options.Keys,
		},
	}

	if len(options.SourcesToUSe) < 1 {
		options.SourcesToUSe = sources.List
	}

	for _, source := range options.SourcesToUSe {
		switch source {
		case "anubis":
			finder.Sources[source] = &anubis.Source{}
		case "bevigil":
			finder.Sources[source] = &bevigil.Source{}
		case "chaos":
			finder.Sources[source] = &chaos.Source{}
		case "commoncrawl":
			finder.Sources[source] = &commoncrawl.Source{}
		case "crtsh":
			finder.Sources[source] = &crtsh.Source{}
		case "fullhunt":
			finder.Sources[source] = &fullhunt.Source{}
		case "github":
			finder.Sources[source] = &github.Source{}
		case "hackertarget":
			finder.Sources[source] = &hackertarget.Source{}
		case "intelx":
			finder.Sources[source] = &intelx.Source{}
		case "otx":
			finder.Sources[source] = &otx.Source{}
		case "shodan":
			finder.Sources[source] = &shodan.Source{}
		case "urlscan":
			finder.Sources[source] = &urlscan.Source{}
		case "wayback":
			finder.Sources[source] = &wayback.Source{}
		}
	}

	for _, source := range options.SourcesToExclude {
		delete(finder.Sources, source)
	}

	return
}
