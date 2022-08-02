package agent

import (
	"sync"

	"github.com/hueristiq/hqsubfind3r/pkg/sources"
	"github.com/hueristiq/hqsubfind3r/pkg/sources/alienvault"
	"github.com/hueristiq/hqsubfind3r/pkg/sources/anubis"
	"github.com/hueristiq/hqsubfind3r/pkg/sources/archiveis"
	"github.com/hueristiq/hqsubfind3r/pkg/sources/bufferover"
	"github.com/hueristiq/hqsubfind3r/pkg/sources/cebaidu"
	"github.com/hueristiq/hqsubfind3r/pkg/sources/certspotterv0"
	"github.com/hueristiq/hqsubfind3r/pkg/sources/chaos"
	"github.com/hueristiq/hqsubfind3r/pkg/sources/commoncrawl"
	"github.com/hueristiq/hqsubfind3r/pkg/sources/crtsh"
	"github.com/hueristiq/hqsubfind3r/pkg/sources/github"
	"github.com/hueristiq/hqsubfind3r/pkg/sources/hackertarget"
	"github.com/hueristiq/hqsubfind3r/pkg/sources/intelx"
	"github.com/hueristiq/hqsubfind3r/pkg/sources/rapiddns"
	"github.com/hueristiq/hqsubfind3r/pkg/sources/riddler"
	"github.com/hueristiq/hqsubfind3r/pkg/sources/sonar"
	"github.com/hueristiq/hqsubfind3r/pkg/sources/sublist3r"
	"github.com/hueristiq/hqsubfind3r/pkg/sources/threatcrowd"
	"github.com/hueristiq/hqsubfind3r/pkg/sources/threatminer"
	"github.com/hueristiq/hqsubfind3r/pkg/sources/urlscan"
	"github.com/hueristiq/hqsubfind3r/pkg/sources/wayback"
	"github.com/hueristiq/hqsubfind3r/pkg/sources/ximcx"
)

type Agent struct {
	sources map[string]sources.Source
}

func New(uses, exclusions []string) *Agent {
	agent := &Agent{
		sources: make(map[string]sources.Source),
	}

	for _, source := range uses {
		switch source {
		case "alienvault":
			agent.sources[source] = &alienvault.Source{}
		case "anubis":
			agent.sources[source] = &anubis.Source{}
		case "archiveis":
			agent.sources[source] = &archiveis.Source{}
		case "bufferover":
			agent.sources[source] = &bufferover.Source{}
		case "cebaidu":
			agent.sources[source] = &cebaidu.Source{}
		case "certspotterv0":
			agent.sources[source] = &certspotterv0.Source{}
		case "chaos":
			agent.sources[source] = &chaos.Source{}
		case "commoncrawl":
			agent.sources[source] = &commoncrawl.Source{}
		case "crtsh":
			agent.sources[source] = &crtsh.Source{}
		case "github":
			agent.sources[source] = &github.Source{}
		case "hackertarget":
			agent.sources[source] = &hackertarget.Source{}
		case "intelx":
			agent.sources[source] = &intelx.Source{}
		case "rapiddns":
			agent.sources[source] = &rapiddns.Source{}
		case "riddler":
			agent.sources[source] = &riddler.Source{}
		case "sonar":
			agent.sources[source] = &sonar.Source{}
		case "sublist3r":
			agent.sources[source] = &sublist3r.Source{}
		case "threatcrowd":
			agent.sources[source] = &threatcrowd.Source{}
		case "threatminer":
			agent.sources[source] = &threatminer.Source{}
		case "urlscan":
			agent.sources[source] = &urlscan.Source{}
		case "wayback":
			agent.sources[source] = &wayback.Source{}
		case "ximcx":
			agent.sources[source] = &ximcx.Source{}
		}
	}

	for _, source := range exclusions {
		delete(agent.sources, source)
	}

	return agent
}

func (agent *Agent) Run(domain string, keys *sources.Keys) chan sources.Subdomain {
	results := make(chan sources.Subdomain)

	go func() {
		defer close(results)

		session, err := sources.NewSession(domain, keys)
		if err != nil {
			return
		}

		wg := new(sync.WaitGroup)

		for source, runner := range agent.sources {
			wg.Add(1)

			go func(source string, runner sources.Source) {
				defer wg.Done()

				for resp := range runner.Run(domain, session) {
					results <- resp
				}
			}(source, runner)
		}

		wg.Wait()
	}()

	return results
}
