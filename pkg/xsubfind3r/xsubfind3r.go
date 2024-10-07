package xsubfind3r

import (
	"strings"
	"sync"

	hqgourl "github.com/hueristiq/hq-go-url"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/anubis"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/bevigil"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/bufferover"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/builtwith"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/certspotter"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/chaos"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/commoncrawl"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/crtsh"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/fullhunt"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/github"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/hackertarget"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/intelx"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/leakix"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/otx"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/securitytrails"
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

	parsed := dp.Parse(domain)

	domain = parsed.Root + "." + parsed.TopLevel

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
					if sResult.Type == sources.ResultSubdomain {
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

type Configuration struct {
	SourcesToUSe     []string
	SourcesToExclude []string
	Keys             sources.Keys
}

var dp = hqgourl.NewDomainParser()

func New(cfg *Configuration) (finder *Finder, err error) {
	finder = &Finder{
		Sources: map[string]sources.Source{},
		SourcesConfiguration: &sources.Configuration{
			Keys: cfg.Keys,
		},
	}

	if len(cfg.SourcesToUSe) < 1 {
		cfg.SourcesToUSe = sources.List
	}

	for index := range cfg.SourcesToUSe {
		source := cfg.SourcesToUSe[index]

		switch source {
		case sources.ANUBIS:
			finder.Sources[source] = &anubis.Source{}
		case sources.BEVIGIL:
			finder.Sources[source] = &bevigil.Source{}
		case sources.BUFFEROVER:
			finder.Sources[source] = &bufferover.Source{}
		case sources.BUILTWITH:
			finder.Sources[source] = &builtwith.Source{}
		case sources.CERTSPOTTER:
			finder.Sources[source] = &certspotter.Source{}
		case sources.CHAOS:
			finder.Sources[source] = &chaos.Source{}
		case sources.COMMONCRAWL:
			finder.Sources[source] = &commoncrawl.Source{}
		case sources.CRTSH:
			finder.Sources[source] = &crtsh.Source{}
		case sources.FULLHUNT:
			finder.Sources[source] = &fullhunt.Source{}
		case sources.GITHUB:
			finder.Sources[source] = &github.Source{}
		case sources.HACKERTARGET:
			finder.Sources[source] = &hackertarget.Source{}
		case sources.INTELLIGENCEX:
			finder.Sources[source] = &intelx.Source{}
		case sources.LEAKIX:
			finder.Sources[source] = &leakix.Source{}
		case sources.OPENTHREATEXCHANGE:
			finder.Sources[source] = &otx.Source{}
		case sources.SECURITYTRAILS:
			finder.Sources[source] = &securitytrails.Source{}
		case sources.SHODAN:
			finder.Sources[source] = &shodan.Source{}
		case sources.URLSCAN:
			finder.Sources[source] = &urlscan.Source{}
		case sources.WAYBACK:
			finder.Sources[source] = &wayback.Source{}
		}
	}

	for index := range cfg.SourcesToExclude {
		source := cfg.SourcesToExclude[index]

		delete(finder.Sources, source)
	}

	return
}
