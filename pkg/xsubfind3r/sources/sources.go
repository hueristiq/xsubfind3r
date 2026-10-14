package sources

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/url"
	"regexp"
	"strings"

	hqgohttp "github.com/hueristiq/hq-lib-http-go"
)

type Source interface {
	Name() (name string)
	Run(ctx context.Context, cfg *Configuration, domain string) <-chan Result
	NeedsKeys() (needs bool)
	UseKeys(keys ...string)
}

type Configuration struct {
	Client    *hqgohttp.Client
	Domain    string
	Extractor *regexp.Regexp
}

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

type Result struct {
	Type   ResultType
	Source string
	Value  string
	Error  error
}

type ResultType int

type Keys []string

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

func CloseBody(body io.ReadCloser) {
	if body == nil {
		return
	}

	_ = body.Close()
}

func CompileExtractor(domain string) (extractor *regexp.Regexp) {
	extractor = regexp.MustCompile(fmt.Sprintf(
		`(?i)(?:^|[^a-z0-9.-])(((?:[a-z0-9](?:[a-z0-9-]*[a-z0-9])?\.)+)%s)`,
		regexp.QuoteMeta(domain),
	))

	return
}

func isHostnameChar(c byte) (is bool) {
	switch {
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '-', c == '.':
		is = true
	}

	return
}

func SanitizeError(err error, secrets ...string) (sanitized error) {
	if err == nil {
		return nil
	}

	text := err.Error()

	for _, secret := range secrets {
		if secret == "" {
			continue
		}

		text = strings.ReplaceAll(text, secret, "REDACTED")
		text = strings.ReplaceAll(text, url.QueryEscape(secret), "REDACTED")
	}

	sanitized = errors.New(text)

	return
}
