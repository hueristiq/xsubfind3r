package sources

// Source is an interface inherited by each source.
type Source interface {
	// Run takes in configuration which includes keys/tokens and other stuff,
	// and domain as arguments.
	Run(config *Configuration, domain string) <-chan Result
	// Name returns the name of the source.
	Name() string
}

type Configuration struct {
	Keys Keys
}

type Keys struct {
	Bevigil  []string `yaml:"bevigil"`
	Chaos    []string `yaml:"chaos"`
	Fullhunt []string `yaml:"fullhunt"`
	GitHub   []string `yaml:"github"`
	Intelx   []string `yaml:"intelx"`
	Shodan   []string `yaml:"shodan"`
	URLScan  []string `yaml:"urlscan"`
}

// Result is a result structure returned by a source.
type Result struct {
	Type   ResultType
	Source string
	Value  string
	Error  error
}

// ResultType is the type of result returned by the source.
type ResultType int

// Types of results returned by the source.
const (
	Subdomain ResultType = iota
	Error
)

var List = []string{
	"anubis",
	"bevigil",
	"chaos",
	"commoncrawl",
	"crtsh",
	"fullhunt",
	"github",
	"hackertarget",
	"intelx",
	"otx",
	"shodan",
	"urlscan",
	"wayback",
}
