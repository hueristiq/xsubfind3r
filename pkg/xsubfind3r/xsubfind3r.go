package xsubfind3r

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	hqgohttp "github.com/hueristiq/hq-go-http"
	hqgohttpheader "github.com/hueristiq/hq-go-http/header"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/anubis"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/bevigil"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/builtwith"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/censys"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/certificatedetails"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/certspotter"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/chaos"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/commoncrawl"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/crtsh"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/driftnet"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/fullhunt"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/github"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/hackertarget"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/intelx"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/leakix"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/leakradar"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/otx"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/securitytrails"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/shodan"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/subdomaincenter"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/urlscan"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/virustotal"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/wayback"
)

type Finder struct {
	sources map[string]sources.Source
}

func (finder *Finder) Find(domain string) (results chan sources.Result) {
	results = make(chan sources.Result)

	configuration := &sources.Configuration{
		Extractor: regexp.MustCompile(fmt.Sprintf(`(?i)(?:((?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+))?(%s)`, regexp.QuoteMeta(domain))),
	}

	go func() {
		defer close(results)

		seen := &sync.Map{}

		wg := &sync.WaitGroup{}

		for _, source := range finder.sources {
			wg.Add(1)

			go func(source sources.Source) {
				defer wg.Done()

				sResults := source.Run(configuration, domain)

				for sResult := range sResults {
					if sResult.Type == sources.ResultSubdomain {
						sResult.Value = strings.ToLower(sResult.Value)
						sResult.Value = strings.ReplaceAll(sResult.Value, "*.", "")
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
	SourcesToUSe     []string
	SourcesToExclude []string
	Keys             map[string]sources.Keys
}

func New(cfg *Configuration) (finder *Finder, err error) {
	finder = &Finder{
		sources: make(map[string]sources.Source),
	}

	cc := hqgohttp.DefaultSprayingClientConfiguration

	cc.Headers = []hqgohttp.Header{}
	cc.Timeout = 1 * time.Hour

	if cfg.Client != nil && cfg.Client.UserAgent != "" {
		cc.Headers = append(cc.Headers, hqgohttp.NewSetHeader(hqgohttpheader.UserAgent.String(), cfg.Client.UserAgent))
	}

	hqgohttp.DefaultClient, err = hqgohttp.NewClient(cc)
	if err != nil {
		err = fmt.Errorf("failed to initialize HTTP client: %w", err)

		return
	}

	if len(cfg.SourcesToUSe) < 1 {
		cfg.SourcesToUSe = sources.List
	}

	for index := range cfg.SourcesToUSe {
		source := cfg.SourcesToUSe[index]

		switch source {
		case sources.ANUBIS:
			finder.sources[source] = anubis.New()
		case sources.BEVIGIL:
			finder.sources[source] = bevigil.New()
		case sources.BUILTWITH:
			finder.sources[source] = builtwith.New()
		case sources.CENSYS:
			finder.sources[source] = censys.New()
		case sources.CERTIFICATEDETAILS:
			finder.sources[source] = certificatedetails.New()
		case sources.CERTSPOTTER:
			finder.sources[source] = certspotter.New()
		case sources.CHAOS:
			finder.sources[source] = chaos.New()
		case sources.COMMONCRAWL:
			finder.sources[source] = commoncrawl.New()
		case sources.DRIFTNET:
			finder.sources[source] = driftnet.New()
		case sources.CRTSH:
			finder.sources[source] = crtsh.New()
		case sources.FULLHUNT:
			finder.sources[source] = fullhunt.New()
		case sources.GITHUB:
			finder.sources[source] = github.New()
		case sources.HACKERTARGET:
			finder.sources[source] = hackertarget.New()
		case sources.INTELLIGENCEX:
			finder.sources[source] = intelx.New()
		case sources.LEAKIX:
			finder.sources[source] = leakix.New()
		case sources.LEAKRADAR:
			finder.sources[source] = leakradar.New()
		case sources.OPENTHREATEXCHANGE:
			finder.sources[source] = otx.New()
		case sources.SECURITYTRAILS:
			finder.sources[source] = securitytrails.New()
		case sources.SHODAN:
			finder.sources[source] = shodan.New()
		case sources.SUBDOMAINCENTER:
			finder.sources[source] = subdomaincenter.New()
		case sources.URLSCAN:
			finder.sources[source] = urlscan.New()
		case sources.VIRUSTOTAL:
			finder.sources[source] = virustotal.New()
		case sources.WAYBACK:
			finder.sources[source] = wayback.New()
		}
	}

	for index := range cfg.SourcesToExclude {
		source := cfg.SourcesToExclude[index]

		delete(finder.sources, source)
	}

	for index := range finder.sources {
		source := finder.sources[index]

		if keys, ok := cfg.Keys[source.Name()]; ok {
			source.UseKeys(keys...)
		}
	}

	return
}
