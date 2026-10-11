package sources

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"
	"regexp"
	"strings"

	hqgohttp "github.com/hueristiq/hq-lib-http-go"
)

// CloseBody closes a response body, discarding the close error: once a
// response body has been fully read, its close error carries no actionable
// information. A nil body is ignored.
func CloseBody(body io.ReadCloser) {
	if body == nil {
		return
	}

	_ = body.Close()
}

// Source is a subdomain enumeration source.
type Source interface {
	Name() (name string)
	Run(ctx context.Context, cfg *Configuration, domain string) <-chan Result
	NeedsKeys() (needs bool)
	UseKeys(keys ...string)
}

// Configuration carries the shared state handed to every source run: the
// HTTP client, the domain being enumerated, and the extractor compiled for
// that domain.
type Configuration struct {
	Client    *hqgohttp.Client
	Domain    string
	Extractor *regexp.Regexp
}

// CompileExtractor compiles the subdomain extractor for domain: the domain
// and any dotted labels before it, preceded by a non-hostname character or
// the start of the input (Go regexes have no lookaround, so the leading
// boundary character is consumed as part of the match; [Configuration.Extract]
// validates the trailing boundary from the match indices).
func CompileExtractor(domain string) (extractor *regexp.Regexp) {
	extractor = regexp.MustCompile(fmt.Sprintf(
		`(?i)(?:^|[^a-z0-9.-])(((?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+)%s)`,
		regexp.QuoteMeta(domain),
	))

	return
}

// Extract returns every subdomain of cfg.Domain found in line. Each match is
// validated against hostname boundaries on both sides and against the
// domain itself, so look-alike domains (notexample.com, sub.example.comm,
// example.com.evil.com) are never returned; the validation is
// case-insensitive and preserves the matched casing.
func (cfg *Configuration) Extract(line string) (subdomains []string) {
	domain := strings.ToLower(cfg.Domain)

	for _, match := range cfg.Extractor.FindAllStringSubmatchIndex(line, -1) {
		if len(match) < 4 || match[2] < 0 {
			continue
		}

		value := line[match[2]:match[3]]

		if match[3] < len(line) && isHostnameChar(line[match[3]]) {
			continue
		}

		if !strings.EqualFold(value, cfg.Domain) && !strings.HasSuffix(strings.ToLower(value), "."+domain) {
			continue
		}

		subdomains = append(subdomains, value)
	}

	return
}

// isHostnameChar reports whether c may appear inside a hostname label.
func isHostnameChar(c byte) (is bool) {
	switch {
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '-', c == '.':
		is = true
	}

	return
}

// Result is a single enumeration outcome: a subdomain or a source error.
type Result struct {
	Type   ResultType
	Source string
	Value  string
	Error  error
}

// ResultType discriminates result payloads.
type ResultType int

// Keys is a set of API keys; sources pick one at random per run.
type Keys []string

// PickRandom returns a uniformly random key from the set.
//
// Returns:
//   - key (string): the selected key.
//   - err (error): non-nil when the set is empty.
func (k Keys) PickRandom() (key string, err error) {
	length := len(k)

	if length == 0 {
		err = errors.New("no keys configured")

		return
	}

	maximum := big.NewInt(int64(length))

	indexBig, err := rand.Int(rand.Reader, maximum)
	if err != nil {
		err = fmt.Errorf("failed to generate random index for key selection: %w", err)

		return
	}

	index := indexBig.Int64()

	key = k[index]

	return
}

const (
	ResultSubdomain ResultType = iota
	ResultError
)

const (
	ANUBIS             = "anubis"
	ARQUIVO            = "arquivo"
	BEVIGIL            = "bevigil"
	BUILTWITH          = "builtwith"
	C99                = "c99"
	CENSYS             = "censys"
	CERTIFICATEDETAILS = "certificatedetails"
	CERTSPOTTER        = "certspotter"
	CHAOS              = "chaos"
	COMMONCRAWL        = "commoncrawl"
	CRTSH              = "crtsh"
	DNSDB              = "dnsdb"
	DNSDUMPSTER        = "dnsdumpster"
	DRIFTNET           = "driftnet"
	FOFA               = "fofa"
	FULLHUNT           = "fullhunt"
	GHOSTARCHIVE       = "ghostarchive"
	GITHUB             = "github"
	GITLAB             = "gitlab"
	HACKERTARGET       = "hackertarget"
	HUDSONROCK         = "hudsonrock"
	INTELLIGENCEX      = "intelx"
	LEAKIX             = "leakix"
	LEAKRADAR          = "leakradar"
	MERKLEMAP          = "merklemap"
	NETLAS             = "netlas"
	ONYPHE             = "onyphe"
	OPENTHREATEXCHANGE = "otx"
	QUAKE              = "quake"
	RAPIDDNS           = "rapiddns"
	RIDDLER            = "riddler"
	SECURITYTRAILS     = "securitytrails"
	SHODAN             = "shodan"
	SUBDOMAINCENTER    = "subdomaincenter"
	THREATMINER        = "threatminer"
	URLSCAN            = "urlscan"
	VIRUSTOTAL         = "virustotal"
	WAYBACK            = "wayback"
	WHOISXMLAPI        = "whoisxmlapi"
	ZOOMEYE            = "zoomeye"
)

var List = []string{
	ANUBIS,
	ARQUIVO,
	BEVIGIL,
	BUILTWITH,
	C99,
	CENSYS,
	CERTIFICATEDETAILS,
	CERTSPOTTER,
	CHAOS,
	COMMONCRAWL,
	CRTSH,
	DNSDB,
	DNSDUMPSTER,
	DRIFTNET,
	FOFA,
	FULLHUNT,
	GHOSTARCHIVE,
	GITHUB,
	GITLAB,
	HACKERTARGET,
	HUDSONROCK,
	INTELLIGENCEX,
	LEAKIX,
	LEAKRADAR,
	MERKLEMAP,
	NETLAS,
	ONYPHE,
	OPENTHREATEXCHANGE,
	QUAKE,
	RAPIDDNS,
	RIDDLER,
	SECURITYTRAILS,
	SHODAN,
	SUBDOMAINCENTER,
	THREATMINER,
	URLSCAN,
	VIRUSTOTAL,
	WAYBACK,
	WHOISXMLAPI,
	ZOOMEYE,
}
