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
	SourcesToUse     []string
	SourcesToExclude []string
	Keys             map[string]sources.Keys
}

var (
	Sources = [...]sources.Source{
		anubis.New(),
		bevigil.New(),
		builtwith.New(),
		censys.New(),
		certificatedetails.New(),
		certspotter.New(),
		chaos.New(),
		commoncrawl.New(),
		driftnet.New(),
		crtsh.New(),
		fullhunt.New(),
		github.New(),
		hackertarget.New(),
		intelx.New(),
		leakix.New(),
		leakradar.New(),
		otx.New(),
		securitytrails.New(),
		shodan.New(),
		subdomaincenter.New(),
		urlscan.New(),
		virustotal.New(),
		wayback.New(),
	}
	NameToSourceMap = make(map[string]sources.Source, len(Sources))
)

func init() {
	for i := range Sources {
		source := Sources[i]

		NameToSourceMap[source.Name()] = source
	}
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

	if len(cfg.SourcesToUse) < 1 {
		cfg.SourcesToUse = sources.List
	}

	for i := range cfg.SourcesToUse {
		source := cfg.SourcesToUse[i]

		s, k := NameToSourceMap[source]
		if !k {
			continue
		}

		finder.sources[source] = s
	}

	for i := range cfg.SourcesToExclude {
		source := cfg.SourcesToExclude[i]

		delete(finder.sources, source)
	}

	for i := range finder.sources {
		source := finder.sources[i]

		if keys, ok := cfg.Keys[source.Name()]; ok {
			source.UseKeys(keys...)
		}
	}

	return
}
