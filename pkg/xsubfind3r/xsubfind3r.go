package xsubfind3r

import (
	"strings"
	"sync"

	hqgourl "github.com/hueristiq/hq-go-url"
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
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/fullhunt"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/github"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/hackertarget"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/intelx"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/leakix"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/otx"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/securitytrails"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/shodan"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/subdomaincenter"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/urlscan"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/virustotal"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/wayback"
)

// Finder is the main structure for managing and executing subdomain enumeration.
// It holds the configured data sources and uses a provided configuration to control the behavior.
//
// Fields:
// - sources: A map of enabled data sources to be used for enumeration.
// - configuration: The settings and API keys required for the sources.
type Finder struct {
	sources       map[string]sources.Source
	configuration *sources.Configuration
}

// Find performs subdomain enumeration for a given domain.
// It uses all the enabled sources and streams the results asynchronously through a channel.
//
// Parameters:
// - domain string: The target domain to find subdomains for.
//
// Returns:
// - results chan sources.Result: A channel that streams the results of type `sources.Result`.
func (finder *Finder) Find(domain string) (results chan sources.Result) {
	results = make(chan sources.Result)

	parsed := dp.Parse(domain)

	domain = parsed.SLD + "." + parsed.TLD

	finder.configuration.Extractor = hqgourl.NewDomainExtractor(
		hqgourl.DomainExtractorWithRootDomainPattern(parsed.SLD),
		hqgourl.DomainExtractorWithTLDPattern(parsed.TLD),
	).CompileRegex()

	go func() {
		defer close(results)

		seen := &sync.Map{}

		wg := &sync.WaitGroup{}

		for _, source := range finder.sources {
			wg.Add(1)

			go func(source sources.Source) {
				defer wg.Done()

				sResults := source.Run(finder.configuration, domain)

				for sResult := range sResults {
					if sResult.Type == sources.ResultSubdomain {
						sResult.Value = strings.ToLower(sResult.Value)
						sResult.Value = strings.ReplaceAll(sResult.Value, "*.", "")

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

// Configuration represents the user-defined settings for the Finder.
// It specifies which sources to use or exclude and includes API keys for external sources.
//
// Fields:
// - SourcesToUse []string: List of sources to be used for enumeration.
// - SourcesToExclude []string: List of sources to be excluded.
// - Keys sources.Keys: API keys for authenticated sources.
type Configuration struct {
	SourcesToUSe     []string
	SourcesToExclude []string
	Keys             sources.Keys
}

// dp is a domain parser used to normalize domains into their root and top-level domain (TLD) components.
var dp = hqgourl.NewDomainParser()

// New creates and initializes a new Finder instance.
// It enables the specified sources, applies exclusions, and sets the required configuration.
//
// Parameters:
// - cfg *Configuration: The configuration specifying sources, exclusions, and API keys.
//
// Returns:
// - finder *Finder: A pointer to the initialized Finder instance.
// - err error: Returns an error if initialization fails, otherwise nil.
func New(cfg *Configuration) (finder *Finder, err error) {
	finder = &Finder{
		sources: map[string]sources.Source{},
		configuration: &sources.Configuration{
			Keys: cfg.Keys,
		},
	}

	if len(cfg.SourcesToUSe) < 1 {
		cfg.SourcesToUSe = sources.List
	}

	for _, source := range cfg.SourcesToUSe {
		switch source {
		case sources.ANUBIS:
			finder.sources[source] = &anubis.Source{}
		case sources.BEVIGIL:
			finder.sources[source] = &bevigil.Source{}
		case sources.BUILTWITH:
			finder.sources[source] = &builtwith.Source{}
		case sources.CENSYS:
			finder.sources[source] = &censys.Source{}
		case sources.CERTIFICATEDETAILS:
			finder.sources[source] = &certificatedetails.Source{}
		case sources.CERTSPOTTER:
			finder.sources[source] = &certspotter.Source{}
		case sources.CHAOS:
			finder.sources[source] = &chaos.Source{}
		case sources.COMMONCRAWL:
			finder.sources[source] = &commoncrawl.Source{}
		case sources.CRTSH:
			finder.sources[source] = &crtsh.Source{}
		case sources.FULLHUNT:
			finder.sources[source] = &fullhunt.Source{}
		case sources.GITHUB:
			finder.sources[source] = &github.Source{}
		case sources.HACKERTARGET:
			finder.sources[source] = &hackertarget.Source{}
		case sources.INTELLIGENCEX:
			finder.sources[source] = &intelx.Source{}
		case sources.LEAKIX:
			finder.sources[source] = &leakix.Source{}
		case sources.OPENTHREATEXCHANGE:
			finder.sources[source] = &otx.Source{}
		case sources.SECURITYTRAILS:
			finder.sources[source] = &securitytrails.Source{}
		case sources.SHODAN:
			finder.sources[source] = &shodan.Source{}
		case sources.SUBDOMAINCENTER:
			finder.sources[source] = &subdomaincenter.Source{}
		case sources.URLSCAN:
			finder.sources[source] = &urlscan.Source{}
		case sources.WAYBACK:
			finder.sources[source] = &wayback.Source{}
		case sources.VIRUSTOTAL:
			finder.sources[source] = &virustotal.Source{}
		}
	}

	for index := range cfg.SourcesToExclude {
		source := cfg.SourcesToExclude[index]

		delete(finder.sources, source)
	}

	return
}
