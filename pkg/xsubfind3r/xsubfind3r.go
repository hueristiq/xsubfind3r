package xsubfind3r

import (
	"context"
	"fmt"
	"strings"
	"sync"
	"time"

	hqgohttp "github.com/hueristiq/hq-lib-http-go"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/anubis"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/arquivo"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/bevigil"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/builtwith"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/c99"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/censys"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/certificatedetails"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/certspotter"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/chaos"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/commoncrawl"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/crtsh"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/dnsdb"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/dnsdumpster"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/driftnet"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/fofa"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/fullhunt"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/ghostarchive"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/github"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/gitlab"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/hackertarget"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/hudsonrock"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/intelx"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/leakix"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/leakradar"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/merklemap"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/netlas"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/onyphe"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/otx"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/quake"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/rapiddns"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/riddler"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/securitytrails"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/shodan"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/subdomaincenter"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/threatminer"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/urlscan"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/virustotal"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/wayback"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/whoisxmlapi"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/zoomeye"
)

type Finder struct {
	client  *hqgohttp.Client
	sources map[string]sources.Source
}

func (finder *Finder) Find(ctx context.Context, domain string) (results chan sources.Result) {
	results = make(chan sources.Result)

	configuration := &sources.Configuration{
		Client:    finder.client,
		Domain:    domain,
		Extractor: sources.CompileExtractor(domain),
	}

	go func() {
		defer close(results)

		seen := &sync.Map{}

		wg := &sync.WaitGroup{}

		for _, source := range finder.sources {
			wg.Add(1)

			go func(source sources.Source) {
				defer wg.Done()

				sResults := source.Run(ctx, configuration, domain)

				for sResult := range sResults {
					if sResult.Type == sources.ResultSubdomain {
						sResult.Value = strings.ToLower(sResult.Value)
						sResult.Value = strings.TrimPrefix(sResult.Value, "*.")
						sResult.Value = strings.TrimPrefix(sResult.Value, ".")

						_, loaded := seen.LoadOrStore(sResult.Value, struct{}{})
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

type ClientConfiguration struct {
	UserAgent string
}

type Configuration struct {
	Client           *ClientConfiguration
	SourcesToUse     []string
	SourcesToExclude []string
	Keys             map[string]sources.Keys
}

type SourceFactory func() sources.Source

var sourceFactories = map[string]SourceFactory{
	sources.ANUBIS:             anubis.New,
	sources.ARQUIVO:            arquivo.New,
	sources.BEVIGIL:            bevigil.New,
	sources.BUILTWITH:          builtwith.New,
	sources.C99:                c99.New,
	sources.CENSYS:             censys.New,
	sources.CERTIFICATEDETAILS: certificatedetails.New,
	sources.CERTSPOTTER:        certspotter.New,
	sources.CHAOS:              chaos.New,
	sources.COMMONCRAWL:        commoncrawl.New,
	sources.CRTSH:              crtsh.New,
	sources.DNSDB:              dnsdb.New,
	sources.DNSDUMPSTER:        dnsdumpster.New,
	sources.DRIFTNET:           driftnet.New,
	sources.FOFA:               fofa.New,
	sources.FULLHUNT:           fullhunt.New,
	sources.GHOSTARCHIVE:       ghostarchive.New,
	sources.GITHUB:             github.New,
	sources.GITLAB:             gitlab.New,
	sources.HACKERTARGET:       hackertarget.New,
	sources.HUDSONROCK:         hudsonrock.New,
	sources.INTELLIGENCEX:      intelx.New,
	sources.LEAKIX:             leakix.New,
	sources.LEAKRADAR:          leakradar.New,
	sources.MERKLEMAP:          merklemap.New,
	sources.NETLAS:             netlas.New,
	sources.ONYPHE:             onyphe.New,
	sources.OPENTHREATEXCHANGE: otx.New,
	sources.QUAKE:              quake.New,
	sources.RAPIDDNS:           rapiddns.New,
	sources.RIDDLER:            riddler.New,
	sources.SECURITYTRAILS:     securitytrails.New,
	sources.SHODAN:             shodan.New,
	sources.SUBDOMAINCENTER:    subdomaincenter.New,
	sources.THREATMINER:        threatminer.New,
	sources.URLSCAN:            urlscan.New,
	sources.VIRUSTOTAL:         virustotal.New,
	sources.WAYBACK:            wayback.New,
	sources.WHOISXMLAPI:        whoisxmlapi.New,
	sources.ZOOMEYE:            zoomeye.New,
}

func SourceFactories() (factories map[string]SourceFactory) {
	factories = make(map[string]SourceFactory, len(sourceFactories))

	for name, factory := range sourceFactories {
		factories[name] = factory
	}

	return
}

func New(cfg *Configuration) (finder *Finder, err error) {
	finder = &Finder{
		sources: make(map[string]sources.Source),
	}

	ofs := []hqgohttp.ClientOptionFunc{
		hqgohttp.WithHTTPClient(hqgohttp.DefaultHTTPClient()),
		hqgohttp.WithTimeout(1 * time.Hour),
		hqgohttp.WithMaxRetries(3),
		hqgohttp.WithRetryWaitMin(1 * time.Second),
		hqgohttp.WithRetryWaitMax(30 * time.Second),
		hqgohttp.WithResponseReadLimit(4096),
	}

	if cfg.Client != nil && cfg.Client.UserAgent != "" {
		ofs = append(ofs, hqgohttp.WithUserAgent(cfg.Client.UserAgent))
	}

	finder.client, err = hqgohttp.NewClient(ofs...)
	if err != nil {
		err = fmt.Errorf("failed to initialize HTTP client: %w", err)

		return
	}

	sourcesToUse := cfg.SourcesToUse

	if len(sourcesToUse) < 1 {
		sourcesToUse = sources.List
	}

	for _, name := range sourcesToUse {
		factory, ok := sourceFactories[name]
		if !ok {
			err = fmt.Errorf("unknown source %q", name)

			return
		}

		finder.sources[name] = factory()
	}

	for _, name := range cfg.SourcesToExclude {
		delete(finder.sources, name)
	}

	for _, source := range finder.sources {
		if keys, ok := cfg.Keys[source.Name()]; ok {
			source.UseKeys(keys...)
		}
	}

	return
}
