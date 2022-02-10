package sources

import (
	"regexp"

	"github.com/valyala/fasthttp"
)

type Subdomain struct {
	Source string
	Value  string
}

type Source interface {
	Run(string, *Session) chan Subdomain
	Name() string
}

type Keys struct {
	Chaos      string   `json:"chaos"`
	GitHub     []string `json:"github"`
	Intelx     string   `json:"intelx"` // unused, add just for the purpose of adding * on listing sources
	IntelXHost string   `json:"intelXHost"`
	IntelXKey  string   `json:"intelXKey"`
}

type Session struct {
	Extractor *regexp.Regexp
	Keys      *Keys
	Client    *fasthttp.Client
}
