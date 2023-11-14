package scraper

import (
	"strings"
	"sync"

	"github.com/hueristiq/xsubfind3r/pkg/scraper/sources"
	"github.com/hueristiq/xsubfind3r/pkg/scraper/sources/anubis"
	"github.com/hueristiq/xsubfind3r/pkg/scraper/sources/bevigil"
	"github.com/hueristiq/xsubfind3r/pkg/scraper/sources/builtwith"
	"github.com/hueristiq/xsubfind3r/pkg/scraper/sources/certspotter"
	"github.com/hueristiq/xsubfind3r/pkg/scraper/sources/chaos"
	"github.com/hueristiq/xsubfind3r/pkg/scraper/sources/commoncrawl"
	"github.com/hueristiq/xsubfind3r/pkg/scraper/sources/crtsh"
	"github.com/hueristiq/xsubfind3r/pkg/scraper/sources/fullhunt"
	"github.com/hueristiq/xsubfind3r/pkg/scraper/sources/github"
	"github.com/hueristiq/xsubfind3r/pkg/scraper/sources/hackertarget"
	"github.com/hueristiq/xsubfind3r/pkg/scraper/sources/intelx"
	"github.com/hueristiq/xsubfind3r/pkg/scraper/sources/leakix"
	"github.com/hueristiq/xsubfind3r/pkg/scraper/sources/otx"
	"github.com/hueristiq/xsubfind3r/pkg/scraper/sources/shodan"
	"github.com/hueristiq/xsubfind3r/pkg/scraper/sources/urlscan"
	"github.com/hueristiq/xsubfind3r/pkg/scraper/sources/wayback"
)

type Scraper struct {
	Sources       map[string]sources.Source
	Configuration *sources.Configuration
}

func (scraper *Scraper) Scrape(domain string) (results chan sources.Result) {
	results = make(chan sources.Result)

	go func() {
		defer close(results)

		seenSubdomains := &sync.Map{}

		wg := &sync.WaitGroup{}

		for _, source := range scraper.Sources {
			wg.Add(1)

			go func(source sources.Source) {
				defer wg.Done()

				sResults := source.Run(scraper.Configuration, domain)

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

func New(options *Options) (scraper *Scraper) {
	scraper = &Scraper{
		Sources: map[string]sources.Source{},
		Configuration: &sources.Configuration{
			Keys: options.Keys,
		},
	}

	if len(options.SourcesToUSe) < 1 {
		options.SourcesToUSe = sources.List
	}

	for index := range options.SourcesToUSe {
		source := options.SourcesToUSe[index]

		switch source {
		case "anubis":
			scraper.Sources[source] = &anubis.Source{}
		case "bevigil":
			scraper.Sources[source] = &bevigil.Source{}
		case "builtwith":
			scraper.Sources[source] = &builtwith.Source{}
		case "certspotter":
			scraper.Sources[source] = &certspotter.Source{}
		case "chaos":
			scraper.Sources[source] = &chaos.Source{}
		case "commoncrawl":
			scraper.Sources[source] = &commoncrawl.Source{}
		case "crtsh":
			scraper.Sources[source] = &crtsh.Source{}
		case "fullhunt":
			scraper.Sources[source] = &fullhunt.Source{}
		case "github":
			scraper.Sources[source] = &github.Source{}
		case "hackertarget":
			scraper.Sources[source] = &hackertarget.Source{}
		case "intelx":
			scraper.Sources[source] = &intelx.Source{}
		case "leakix":
			scraper.Sources[source] = &leakix.Source{}
		case "otx":
			scraper.Sources[source] = &otx.Source{}
		case "shodan":
			scraper.Sources[source] = &shodan.Source{}
		case "urlscan":
			scraper.Sources[source] = &urlscan.Source{}
		case "wayback":
			scraper.Sources[source] = &wayback.Source{}
		}
	}

	for index := range options.SourcesToExclude {
		source := options.SourcesToExclude[index]

		delete(scraper.Sources, source)
	}

	return
}
