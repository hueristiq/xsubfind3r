package sources

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"io"
	"math/big"
	"net/http"
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
	ResultWarning
)

const MaxPages = 100

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
		value, ok := trimEncodedSlashPrefix(text, match[2], text[match[2]:match[3]])
		if !ok {
			continue
		}

		end := match[3]

		if end < len(text) && isHostLabelCharacter(text[end]) {
			if text[end] != '.' || (end+1 < len(text) && isHostLabelCharacter(text[end+1])) {
				continue
			}
		}

		subdomains = append(subdomains, value)
	}

	return
}

// trimEncodedSlashPrefix removes the "2f" left behind when a match starts at
// the percent-encoded slash in "%2F". The match must be directly preceded by
// '%'; a remainder that is empty or starts with a dot is not a recoverable
// hostname and is dropped.
func trimEncodedSlashPrefix(text string, start int, value string) (trimmed string, ok bool) {
	if start == 0 || text[start-1] != '%' || len(value) < 3 || !strings.EqualFold(value[:2], "2f") {
		return value, true
	}

	trimmed = value[2:]

	if trimmed == "" || trimmed[0] == '.' {
		return "", false
	}

	return trimmed, true
}

func isHostLabelCharacter(character byte) (ok bool) {
	switch {
	case character >= 'a' && character <= 'z', character >= 'A' && character <= 'Z', character >= '0' && character <= '9', character == '.', character == '-', character == '_':
		ok = true
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

func CheckResponse(res *http.Response, secrets ...string) (err error) {
	if res == nil {
		err = errors.New("empty response")

		return
	}

	if res.StatusCode >= http.StatusOK && res.StatusCode < http.StatusMultipleChoices {
		return
	}

	snippet := ""

	if res.Body != nil {
		var body []byte

		body, _ = io.ReadAll(io.LimitReader(res.Body, 4*1024))
		snippet = strings.TrimSpace(string(body))
	}

	if snippet == "" {
		err = fmt.Errorf("unexpected status %s", res.Status)
	} else {
		err = fmt.Errorf("unexpected status %s: %s", res.Status, snippet)
	}

	err = RedactError(err, secrets...)

	return
}

func ErrorResult(source string, err error) (result Result) {
	result = Result{
		Type:   ResultError,
		Source: source,
		Error:  err,
	}

	return
}

func WarningResult(source, value string) (result Result) {
	result = Result{
		Type:   ResultWarning,
		Source: source,
		Value:  value,
	}

	return
}

func SubdomainResult(source, value string) (result Result) {
	result = Result{
		Type:   ResultSubdomain,
		Source: source,
		Value:  value,
	}

	return
}
