package sources

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"net/url"
	"regexp"
	"strings"

	hqgohttp "github.com/hueristiq/hq-lib-http-go"
)

type redactedError struct {
	err     error
	message string
}

func (e *redactedError) Error() (message string) {
	return e.message
}

func (e *redactedError) Unwrap() (err error) {
	return e.err
}

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
	ResultSubdomain ResultType = iota + 1
	ResultError
)

const (
	Anubis             = "anubis"
	Arquivo            = "arquivo"
	Bevigil            = "bevigil"
	Builtwith          = "builtwith"
	C99                = "c99"
	Censys             = "censys"
	Certificatedetails = "certificatedetails"
	Certspotter        = "certspotter"
	Chaos              = "chaos"
	Commoncrawl        = "commoncrawl"
	Crtsh              = "crtsh"
	DNSDB              = "dnsdb"
	DNSDumpster        = "dnsdumpster"
	Driftnet           = "driftnet"
	Fofa               = "fofa"
	Fullhunt           = "fullhunt"
	Ghostarchive       = "ghostarchive"
	Github             = "github"
	Gitlab             = "gitlab"
	Hackertarget       = "hackertarget"
	Hudsonrock         = "hudsonrock"
	IntelligenceX      = "intelx"
	Leakix             = "leakix"
	Leakradar          = "leakradar"
	Merklemap          = "merklemap"
	Netlas             = "netlas"
	Onyphe             = "onyphe"
	OpenThreatExchange = "otx"
	Quake              = "quake"
	RapidDNS           = "rapiddns"
	Riddler            = "riddler"
	Securitytrails     = "securitytrails"
	Shodan             = "shodan"
	Subdomaincenter    = "subdomaincenter"
	Threatminer        = "threatminer"
	URLScan            = "urlscan"
	Virustotal         = "virustotal"
	Wayback            = "wayback"
	WhoisXMLAPI        = "whoisxmlapi"
	Zoomeye            = "zoomeye"
)

var List = []string{
	Anubis,
	Arquivo,
	Bevigil,
	Builtwith,
	C99,
	Censys,
	Certificatedetails,
	Certspotter,
	Chaos,
	Commoncrawl,
	Crtsh,
	DNSDB,
	DNSDumpster,
	Driftnet,
	Fofa,
	Fullhunt,
	Ghostarchive,
	Github,
	Gitlab,
	Hackertarget,
	Hudsonrock,
	IntelligenceX,
	Leakix,
	Leakradar,
	Merklemap,
	Netlas,
	Onyphe,
	OpenThreatExchange,
	Quake,
	RapidDNS,
	Riddler,
	Securitytrails,
	Shodan,
	Subdomaincenter,
	Threatminer,
	URLScan,
	Virustotal,
	Wayback,
	WhoisXMLAPI,
	Zoomeye,
}

var _ error = (*redactedError)(nil)

func RedactError(err error, secrets ...string) (redacted error) {
	if err == nil || len(secrets) == 0 {
		return err
	}

	message := err.Error()

	for _, secret := range secrets {
		if secret == "" {
			continue
		}

		message = strings.ReplaceAll(message, secret, "[REDACTED]")
		message = strings.ReplaceAll(message, url.QueryEscape(secret), "[REDACTED]")
	}

	if message == err.Error() {
		return err
	}

	return &redactedError{
		err:     err,
		message: message,
	}
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
