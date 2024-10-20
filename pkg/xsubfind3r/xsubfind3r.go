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

// Finder is the main structure that manages the interaction with OSINT sources.
// It holds the available data sources and the configuration used for searching.
type Finder struct {
	// sources is a map of source names to their corresponding implementations.
	// Each source implements the Source interface, which allows domain searches.
	sources map[string]sources.Source
	// configuration contains configuration options such as API keys
	// and other settings needed by the data sources.
	configuration *sources.Configuration
}

// Find takes a domain name and starts the subdomain search process across all
// the sources specified in the configuration. It returns a channel through which
// the search results (of type Result) are streamed asynchronously.
func (finder *Finder) Find(domain string) (results chan sources.Result) {
	// Initialize the results channel where subdomain findings are sent.
	results = make(chan sources.Result)

	// Parse the given domain using a domain parser.
	parsed := dp.Parse(domain)

	// Rebuild the domain as "root.tld" format.
	domain = parsed.Root + "." + parsed.TopLevel

	finder.configuration.Extractor = hqgourl.NewDomainExtractor(
		hqgourl.DomainExtractorWithRootDomainPattern(parsed.Root),
		hqgourl.DomainExtractorWithTLDPattern(parsed.TopLevel),
	).CompileRegex()

	// Launch a goroutine to perform the search concurrently across all sources.
	go func() {
		// Ensure the results channel is closed once all search operations complete.
		defer close(results)

		// A thread-safe map to store already-seen subdomains, avoiding duplicates.
		seenSubdomains := &sync.Map{}

		// WaitGroup ensures all source goroutines finish before exiting.
		wg := &sync.WaitGroup{}

		// Iterate over all the sources in the Finder.
		for _, source := range finder.sources {
			wg.Add(1)

			// Start a new goroutine for each source to fetch subdomains concurrently.
			go func(source sources.Source) {
				// Decrement the WaitGroup counter when this goroutine completes.
				defer wg.Done()

				// Call the source's Run method to start the subdomain search.
				sResults := source.Run(finder.configuration, domain)

				// Process each result as it's received from the source.
				for sResult := range sResults {
					// If the result is a subdomain, process it.
					if sResult.Type == sources.ResultSubdomain {
						// Convert the subdomain to lowercase and strip any wildcards (e.g., "*.")
						sResult.Value = strings.ToLower(sResult.Value)
						sResult.Value = strings.ReplaceAll(sResult.Value, "*.", "")

						// Check if the subdomain has already been seen using sync.Map.
						_, loaded := seenSubdomains.LoadOrStore(sResult.Value, struct{}{})
						if loaded {
							// If the subdomain is already in the map, skip it.
							continue
						}
					}

					// Send the result down the results channel.
					results <- sResult
				}
			}(source)
		}

		// Wait for all goroutines to finish before exiting.
		wg.Wait()
	}()

	// Return the channel that will stream subdomain results.
	return
}

// Configuration holds the configuration for Finder, including
// the sources to use, sources to exclude, and the necessary API keys.
type Configuration struct {
	// SourcesToUse is a list of source names that should be used for the search.
	SourcesToUSe []string
	// SourcesToExclude is a list of source names that should be excluded from the search.
	SourcesToExclude []string
	// Keys contains the API keys for each data source.
	Keys sources.Keys
}

// dp is a domain parser used to normalize domains into their root and top-level domain (TLD) components.
var dp = hqgourl.NewDomainParser()

// New creates a new Finder instance based on the provided Configuration.
// It initializes the Finder with the selected sources and ensures that excluded sources are not used.
func New(cfg *Configuration) (finder *Finder, err error) {
	// Initialize a Finder instance with an empty map of sources and the provided configuration.
	finder = &Finder{
		sources: map[string]sources.Source{},
		configuration: &sources.Configuration{
			Keys: cfg.Keys,
		},
	}

	// If no specific sources are provided, use the default list of all sources.
	if len(cfg.SourcesToUSe) < 1 {
		cfg.SourcesToUSe = sources.List
	}

	// Loop through the selected sources and initialize each one.
	for _, source := range cfg.SourcesToUSe {
		// Depending on the source name, initialize the appropriate source and add it to the map.
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

	// Remove any sources that are specified in the SourcesToExclude list.
	for index := range cfg.SourcesToExclude {
		source := cfg.SourcesToExclude[index]

		delete(finder.sources, source)
	}

	// Return the Finder instance with all the selected sources.
	return
}
