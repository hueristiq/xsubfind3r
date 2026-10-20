package xsubfind3r

import (
	"context"
	"fmt"
	"hash/fnv"
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

		deduper := newDeduper()

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

						if sResult.Value != domain && !strings.HasSuffix(sResult.Value, "."+domain) {
							continue
						}

						if deduper.seen(sResult.Value) {
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

var registry = map[string]func() sources.Source{
	sources.Anubis:             anubis.New,
	sources.Arquivo:            arquivo.New,
	sources.Bevigil:            bevigil.New,
	sources.Builtwith:          builtwith.New,
	sources.C99:                c99.New,
	sources.Censys:             censys.New,
	sources.Certificatedetails: certificatedetails.New,
	sources.Certspotter:        certspotter.New,
	sources.Chaos:              chaos.New,
	sources.Commoncrawl:        commoncrawl.New,
	sources.Crtsh:              crtsh.New,
	sources.DNSDB:              dnsdb.New,
	sources.DNSDumpster:        dnsdumpster.New,
	sources.Driftnet:           driftnet.New,
	sources.Fofa:               fofa.New,
	sources.Fullhunt:           fullhunt.New,
	sources.Ghostarchive:       ghostarchive.New,
	sources.Github:             github.New,
	sources.Gitlab:             gitlab.New,
	sources.Hackertarget:       hackertarget.New,
	sources.Hudsonrock:         hudsonrock.New,
	sources.IntelligenceX:      intelx.New,
	sources.Leakix:             leakix.New,
	sources.Leakradar:          leakradar.New,
	sources.Merklemap:          merklemap.New,
	sources.Netlas:             netlas.New,
	sources.Onyphe:             onyphe.New,
	sources.OpenThreatExchange: otx.New,
	sources.Quake:              quake.New,
	sources.RapidDNS:           rapiddns.New,
	sources.Riddler:            riddler.New,
	sources.Securitytrails:     securitytrails.New,
	sources.Shodan:             shodan.New,
	sources.Subdomaincenter:    subdomaincenter.New,
	sources.Threatminer:        threatminer.New,
	sources.URLScan:            urlscan.New,
	sources.Virustotal:         virustotal.New,
	sources.Wayback:            wayback.New,
	sources.WhoisXMLAPI:        whoisxmlapi.New,
	sources.Zoomeye:            zoomeye.New,
}

// NewSource returns a fresh instance of the source registered under
// name.
//
// Parameters:
//   - name (string): the registered source name.
//
// Returns:
//   - source (sources.Source): the new source instance; nil when name
//     is unknown.
//   - ok (bool): whether name is registered.
func NewSource(name string) (source sources.Source, ok bool) {
	constructor, exists := registry[name]
	if !exists {
		return nil, false
	}

	return constructor(), true
}

// New returns a finder configured with cfg. Every finder owns fresh
// source instances, so finders are safe to construct and use
// concurrently.
//
// Parameters:
//   - cfg (*Configuration): the finder configuration.
//
// Returns:
//   - finder (*Finder): the configured finder.
//   - err (error): non-nil when the HTTP client cannot be initialized
//     or a requested source name is unknown.
func New(cfg *Configuration) (finder *Finder, err error) {
	finder = &Finder{
		sources: make(map[string]sources.Source, len(sources.List)),
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

	for i := range sourcesToUse {
		name := sourcesToUse[i]

		source, ok := NewSource(name)
		if !ok {
			err = fmt.Errorf("unknown source %q; supported sources: %s", name, strings.Join(sources.List, ", "))

			return
		}

		finder.sources[name] = source
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

const dedupShardCount = 32

type deduper struct {
	shards [dedupShardCount]dedupShard
}

type dedupShard struct {
	mu      sync.Mutex
	entries map[uint64]struct{}
}

func newDeduper() (d *deduper) {
	d = &deduper{}

	for i := range d.shards {
		d.shards[i].entries = make(map[uint64]struct{})
	}

	return
}

func (d *deduper) seen(value string) (duplicate bool) {
	hash := dedupHash(value)

	shard := &d.shards[hash%dedupShardCount]

	shard.mu.Lock()
	defer shard.mu.Unlock()

	if _, exists := shard.entries[hash]; exists {
		return true
	}

	shard.entries[hash] = struct{}{}

	return false
}

func dedupHash(value string) (hash uint64) {
	hasher := fnv.New64a()

	_, _ = hasher.Write([]byte(value))

	hash = hasher.Sum64()

	return
}
