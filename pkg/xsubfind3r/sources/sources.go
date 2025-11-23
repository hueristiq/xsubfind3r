package sources

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"regexp"
)

type Source interface {
	Name() (name string)
	Run(cfg *Configuration, domain string) <-chan Result
	UseKeys(keys ...string)
}

type Configuration struct {
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
	BEVIGIL            = "bevigil"
	BUILTWITH          = "builtwith"
	CENSYS             = "censys"
	CERTIFICATEDETAILS = "certificatedetails"
	CERTSPOTTER        = "certspotter"
	CHAOS              = "chaos"
	COMMONCRAWL        = "commoncrawl"
	CRTSH              = "crtsh"
	DRIFTNET           = "driftnet"
	FULLHUNT           = "fullhunt"
	GITHUB             = "github"
	HACKERTARGET       = "hackertarget"
	INTELLIGENCEX      = "intelx"
	LEAKIX             = "leakix"
	LEAKRADAR          = "leakradar"
	OPENTHREATEXCHANGE = "otx"
	SECURITYTRAILS     = "securitytrails"
	SHODAN             = "shodan"
	SUBDOMAINCENTER    = "subdomaincenter"
	URLSCAN            = "urlscan"
	VIRUSTOTAL         = "virustotal"
	WAYBACK            = "wayback"
)

var List = []string{
	ANUBIS,
	BEVIGIL,
	BUILTWITH,
	CENSYS,
	CERTIFICATEDETAILS,
	CERTSPOTTER,
	CHAOS,
	COMMONCRAWL,
	CRTSH,
	DRIFTNET,
	FULLHUNT,
	GITHUB,
	HACKERTARGET,
	INTELLIGENCEX,
	LEAKIX,
	LEAKRADAR,
	OPENTHREATEXCHANGE,
	SECURITYTRAILS,
	SHODAN,
	SUBDOMAINCENTER,
	URLSCAN,
	VIRUSTOTAL,
	WAYBACK,
}
