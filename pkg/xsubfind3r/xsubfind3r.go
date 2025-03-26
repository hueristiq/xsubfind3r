// Package xsubfind3r provides the core functionality for performing subdomain
// discovery using multiple data sources. It integrates various sources that implement
// the sources.Source interface, coordinates concurrent subdomain enumeration, and
// aggregates the results.
//
// The package defines a Finder type, which manages enabled sources and configuration
// settings, and provides a Find method to initiate subdomain discovery for a given domain.
// It also defines a Configuration type for user-defined settings and API keys, and
// initializes HTTP client configurations for reliable network requests.
package xsubfind3r

import (
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

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
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/otx"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/securitytrails"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/shodan"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/subdomaincenter"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/urlscan"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/virustotal"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources/wayback"
	hqgohttp "go.source.hueristiq.com/http"
	"go.source.hueristiq.com/url/parser"
)

// Finder is the primary structure for performing subdomain discovery.
// It manages data sources and configuration settings.
//
// Fields:
//   - sources (map[string]sources.Source): A map of string keys to sources.Source interfaces representing the enabled enumeration sources.
//   - configuration (*sources.Configuration): A pointer to the sources.Configuration struct containing API keys and other settings.
type Finder struct {
	sources       map[string]sources.Source
	configuration *sources.Configuration
}

// Find initiates the subdomain discovery process for a specific domain.
// It normalizes the domain name, applies source-specific logic, and streams results via a channel.
// The method uses all enabled sources concurrently and aggregates their results.
//
// Parameters:
//   - domain (string): The target domain for subdomain discovery.
//
// Returns:
//   - results (chan sources.Result): A channel that streams subdomain enumeration results.
func (finder *Finder) Find(domain string) (results chan sources.Result) {
	results = make(chan sources.Result)

	parsed, _ := up.Parse(domain)

	domain = parsed.Domain.SecondLevelDomain + "." + parsed.Domain.TopLevelDomain

	pattern := fmt.Sprintf(`(?i)(?:((?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+))?(%s)`, regexp.QuoteMeta(domain))

	finder.configuration.Extractor = regexp.MustCompile(pattern)

	go func() {
		defer close(results)

		seen := &sync.Map{}

		wg := &sync.WaitGroup{}

		for _, source := range finder.sources {
			wg.Add(1)

			go func(source sources.Source) {
				defer wg.Done()

				sResults := source.Run(domain, finder.configuration)

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
//   - SourcesToUSe ([]string): List of source names to be used for enumeration.
//   - SourcesToExclude ([]string): List of source names to be excluded from enumeration.
//   - Keys (sources.Keys): API keys for authenticated sources.
type Configuration struct {
	SourcesToUSe     []string
	SourcesToExclude []string
	Keys             sources.Keys
}

// dp is a domain parser used to normalize domains into their root and top-level domain (TLD) components.
var up = parser.New(
	parser.WithDefaultScheme("http"),
)

func init() {
	cfg := hqgohttp.DefaultSprayingClientConfiguration

	cfg.Timeout = 1 * time.Hour

	hqgohttp.DefaultClient, _ = hqgohttp.NewClient(cfg)
}

// New initializes a new Finder instance with the specified configuration.
// It sets up the enabled sources, applies exclusions, and configures the Finder.
//
// Parameters:
//   - cfg (*Configuration): The user-defined configuration for sources and API keys.
//
// Returns:
//   - finder (*Finder): A pointer to the initialized Finder instance.
//   - err (error): An error object if initialization fails, or nil on success.
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
		case sources.DRIFTNET:
			finder.sources[source] = &driftnet.Source{}
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
