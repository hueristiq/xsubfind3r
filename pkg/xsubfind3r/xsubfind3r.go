package xsubfind3r

import (
	"context"
	"errors"
	"fmt"
	"hash/fnv"
	"net/http"
	"regexp"
	"strings"
	"sync"
	"time"

	hqgohttp "github.com/hueristiq/hq-lib-http-go"
	hqgohttpheader "github.com/hueristiq/hq-lib-http-go/header"
	hqgoretrier "github.com/hueristiq/hq-lib-retrier-go"
	hqgoretrierbackoff "github.com/hueristiq/hq-lib-retrier-go/backoff"

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
	results = make(chan sources.Result, 1)

	domain = strings.ToLower(domain)

	extractor, err := compileExtractor(domain)
	if err != nil {
		results <- sources.ErrorResult("xsubfind3r", err)

		close(results)

		return results
	}

	configuration := &sources.Configuration{
		Client:    finder.client,
		Extractor: extractor,
		Validator: func(target string) (subdomain string, valid bool) {
			subdomain, valid = validateSubdomain(target, domain)

			return
		},
	}

	go func() {
		defer close(results)

		deduper := newDeduper()

		wg := &sync.WaitGroup{}

		for _, source := range finder.sources {
			wg.Add(1)

			go func(source sources.Source) {
				defer wg.Done()

				relayResults(ctx, source.Run(ctx, configuration, domain), deduper, results)
			}(source)
		}

		wg.Wait()
	}()

	return
}

func relayResults(ctx context.Context, sourceResults <-chan sources.Result, deduper *deduper, results chan<- sources.Result) {
	for {
		select {
		case <-ctx.Done():
			drainResults(sourceResults)

			return
		case result, ok := <-sourceResults:
			if !ok {
				return
			}

			if result.Type == sources.ResultSubdomain && deduper.seen(result.Value) {
				continue
			}

			select {
			case results <- result:
			case <-ctx.Done():
				drainResults(sourceResults)

				return
			}
		}
	}
}

func drainResults(sourceResults <-chan sources.Result) {
	for range sourceResults {
	}
}

type deduper struct {
	shards [dedupShardCount]dedupShard
}

func (d *deduper) seen(value string) (duplicate bool) {
	hash := dedupHash(value)

	shard := &d.shards[hash%dedupShardCount]

	shard.mu.Lock()
	defer shard.mu.Unlock()

	if _, exists := shard.entries[value]; exists {
		return true
	}

	shard.entries[value] = struct{}{}

	return false
}

type dedupShard struct {
	mu      sync.Mutex
	entries map[string]struct{}
}

type ClientConfiguration struct {
	UserAgent string
}

type Configuration struct {
	Client           *ClientConfiguration
	SourcesToUse     []string
	SourcesToExclude []string
	Keys             map[string]sources.Keys
	Timeout          time.Duration
}

const (
	dedupShardCount = 32

	defaultTimeout = 30 * time.Second
)

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

func New(cfg *Configuration) (finder *Finder, err error) {
	if cfg == nil {
		err = errors.New("configuration is nil")

		return
	}

	finder = &Finder{
		sources: make(map[string]sources.Source, len(sources.List)),
	}

	timeout := cfg.Timeout
	if timeout <= 0 {
		timeout = defaultTimeout
	}

	ofs := []hqgohttp.ClientOptionFunc{
		hqgohttp.WithHTTPClient(hqgohttp.DefaultHTTPPooledClient()),
		hqgohttp.WithTimeout(timeout),
		hqgohttp.WithRetrier(newRetrier()),
		hqgohttp.WithResponseReadLimit(4096),
	}

	if cfg.Client != nil && cfg.Client.UserAgent != "" {
		ofs = append(ofs, hqgohttp.WithHeader(hqgohttpheader.UserAgent.String(), cfg.Client.UserAgent))
	}

	finder.client = hqgohttp.New(ofs...)

	sourcesToUse := cfg.SourcesToUse

	if len(sourcesToUse) < 1 {
		sourcesToUse = sources.List
	}

	for i := range sourcesToUse {
		name := sourcesToUse[i]

		source, ok := NewSource(name)
		if !ok {
			err = fmt.Errorf("unknown source %q", name)

			return
		}

		finder.sources[name] = source
	}

	for _, name := range cfg.SourcesToExclude {
		if _, ok := registry[name]; !ok {
			err = fmt.Errorf("unknown source %q in sources to exclude", name)

			return
		}

		delete(finder.sources, name)
	}

	if len(finder.sources) == 0 {
		err = errors.New("no sources left after applying exclusions")

		return
	}

	for _, source := range finder.sources {
		if keys, ok := cfg.Keys[source.Name()]; ok {
			source.UseKeys(keys...)
		}
	}

	return
}

func NewSource(name string) (source sources.Source, ok bool) {
	constructor, exists := registry[name]
	if !exists {
		return nil, false
	}

	return constructor(), true
}

func compileExtractor(domain string) (extractor *regexp.Regexp, err error) {
	extractor, err = regexp.Compile(fmt.Sprintf(
		`(?i)(?:^|[^a-z0-9_.-])(((?:[a-z0-9_](?:[a-z0-9_-]*[a-z0-9_])?\.)+)%s)`,
		regexp.QuoteMeta(domain),
	))
	if err != nil {
		return nil, fmt.Errorf("failed to compile subdomain extractor: %w", err)
	}

	return
}

func validateSubdomain(target, domain string) (subdomain string, valid bool) {
	subdomain = strings.ToLower(strings.TrimSpace(target))
	subdomain = strings.TrimPrefix(subdomain, "*.")
	subdomain = strings.TrimSuffix(subdomain, ".")

	if !validHostname(subdomain) {
		return subdomain, false
	}

	if subdomain == domain || !strings.HasSuffix(subdomain, "."+domain) {
		return subdomain, false
	}

	return subdomain, true
}

func validHostname(hostname string) (valid bool) {
	if hostname == "" || len(hostname) > 253 {
		return false
	}

	for _, label := range strings.Split(hostname, ".") {
		if label == "" || len(label) > 63 {
			return false
		}

		for index := range len(label) {
			if !isHostnameChar(label[index]) {
				return false
			}
		}

		if !isHostnameEdge(label[0]) || !isHostnameEdge(label[len(label)-1]) {
			return false
		}
	}

	return true
}

func isHostnameChar(char byte) (valid bool) {
	switch {
	case char >= 'a' && char <= 'z',
		char >= '0' && char <= '9',
		char == '-',
		char == '_':
		valid = true
	}

	return
}

func isHostnameEdge(char byte) (valid bool) {
	switch {
	case char >= 'a' && char <= 'z',
		char >= '0' && char <= '9',
		char == '_':
		valid = true
	}

	return
}

func newDeduper() (d *deduper) {
	d = &deduper{}

	for i := range d.shards {
		d.shards[i].entries = make(map[string]struct{})
	}

	return
}

func dedupHash(value string) (hash uint64) {
	hasher := fnv.New64a()

	_, _ = hasher.Write([]byte(value))

	hash = hasher.Sum64()

	return
}

func newRetrier() (retrier hqgohttp.Retrier) {
	retrier = hqgohttp.RetrierFunc(func(ctx context.Context, operation hqgohttp.RetryOperation, options hqgohttp.RetrierOptions) (*http.Response, error) {
		return hqgoretrier.RetryWithData(ctx, operation,
			hqgoretrier.WithMaxAttempts(4),
			hqgoretrier.WithWaitMin(1*time.Second),
			hqgoretrier.WithWaitMax(30*time.Second),
			hqgoretrier.WithBackoff(func(minDelay, maxDelay time.Duration) hqgoretrierbackoff.Backoff {
				base := hqgoretrierbackoff.ExponentialWithDecorrelatedJitter(minDelay, maxDelay)

				return func(attempt int) time.Duration {
					if wait, ok := options.RetryAfter(); ok {
						return min(wait, maxDelay)
					}

					return base(attempt)
				}
			}),
			hqgoretrier.WithRetryOn(options.RetryOn),
			hqgoretrier.WithOnRetry(func(event hqgoretrier.RetryEvent) {
				options.OnRetry(hqgohttp.RetrierEvent{Attempt: event.Attempt, Err: event.Err, Wait: event.Wait})
			}),
		)
	})

	return
}
