package xsubfind3r

import (
	"regexp"
	"sync"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/alienvault"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/anubis"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/archiveis"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/bevigil"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/bufferover"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/cebaidu"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/censys"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/certspotterv0"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/chaos"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/commoncrawl"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/crtsh"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/fullhunt"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/github"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/hackertarget"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/intelx"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/rapiddns"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/riddler"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/sonar"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/sublist3r"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/threatcrowd"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/threatminer"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/urlscan"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/wayback"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/ximcx"
)

type Options struct {
	Domain           string
	SourcesToExclude []string
	SourcesToUSe     []string
	Keys             sources.Keys
}

type Finder struct {
	Sources              map[string]sources.Source
	SourcesConfiguration *sources.Configuration
}

func New(options *Options) (finder *Finder) {
	finder = &Finder{
		Sources: map[string]sources.Source{},
		SourcesConfiguration: &sources.Configuration{
			Keys:            options.Keys,
			Domain:          options.Domain,
			SubdomainsRegex: regexp.MustCompile(`[a-zA-Z0-9\*_.-]+\.` + options.Domain),
		},
	}

	if len(options.SourcesToUSe) < 1 {
		options.SourcesToUSe = sources.List
	}

	for _, source := range options.SourcesToUSe {
		switch source {
		case "alienvault":
			finder.Sources[source] = &alienvault.Source{}
		case "anubis":
			finder.Sources[source] = &anubis.Source{}
		case "archiveis":
			finder.Sources[source] = &archiveis.Source{}
		case "bevigil":
			finder.Sources[source] = &bevigil.Source{}
		case "bufferover":
			finder.Sources[source] = &bufferover.Source{}
		case "cebaidu":
			finder.Sources[source] = &cebaidu.Source{}
		case "censys":
			finder.Sources[source] = &censys.Source{}
		case "certspotterv0":
			finder.Sources[source] = &certspotterv0.Source{}
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
		case "rapiddns":
			finder.Sources[source] = &rapiddns.Source{}
		case "riddler":
			finder.Sources[source] = &riddler.Source{}
		case "sonar":
			finder.Sources[source] = &sonar.Source{}
		case "sublist3r":
			finder.Sources[source] = &sublist3r.Source{}
		case "threatcrowd":
			finder.Sources[source] = &threatcrowd.Source{}
		case "threatminer":
			finder.Sources[source] = &threatminer.Source{}
		case "urlscan":
			finder.Sources[source] = &urlscan.Source{}
		case "wayback":
			finder.Sources[source] = &wayback.Source{}
		case "ximcx":
			finder.Sources[source] = &ximcx.Source{}
		}
	}

	for _, source := range options.SourcesToExclude {
		delete(finder.Sources, source)
	}

	return
}

func (finder *Finder) Find() (subdomains chan sources.Subdomain) {
	subdomains = make(chan sources.Subdomain)

	go func() {
		defer close(subdomains)

		wg := &sync.WaitGroup{}
		seen := &sync.Map{}

		for name := range finder.Sources {
			wg.Add(1)

			go func(source sources.Source) {
				defer wg.Done()

				for subdomain := range source.Run(finder.SourcesConfiguration) {
					value := subdomain.Value

					if value == "" {
						return
					}

					_, loaded := seen.LoadOrStore(value, struct{}{})
					if loaded {
						continue
					}

					subdomains <- subdomain
				}
			}(finder.Sources[name])
		}

		wg.Wait()
	}()

	return
}
