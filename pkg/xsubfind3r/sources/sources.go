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

type Configuration struct {
	Client    *hqgohttp.Client
	Extractor *regexp.Regexp
	Validator func(target string) (subdomain string, valid bool)
}

type ResultType int

type Result struct {
	Type   ResultType
	Source string
	Value  string
	Error  error
}

type Source interface {
	Name() (name string)
	NeedsKeys() (needs bool)
	UseKeys(keys ...string)
	Run(ctx context.Context, cfg *Configuration, domain string) <-chan Result
}

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

	if key == "" {
		err = errors.New("selected key is empty")

		return
	}

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

func ExtractSubdomains(extractor *regexp.Regexp, text string) (subdomains []string) {
	subdomains = make([]string, 0)

	for _, match := range extractor.FindAllStringSubmatchIndex(text, -1) {
		value := text[match[2]:match[3]]
		end := match[3]

		if end < len(text) && isHostnameChar(text[end]) {
			if text[end] != '.' || (end+1 < len(text) && isHostnameChar(text[end+1])) {
				continue
			}
		}

		subdomains = append(subdomains, value)
	}

	return
}

func isHostnameChar(c byte) (is bool) {
	switch {
	case c >= 'a' && c <= 'z', c >= 'A' && c <= 'Z', c >= '0' && c <= '9', c == '.', c == '-', c == '_':
		is = true
	}

	return
}

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
