package sources

import (
	"context"
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"regexp"

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
	Extractor *regexp.Regexp
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

	var indexBig *big.Int

	indexBig, err = rand.Int(rand.Reader, maximum)
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
