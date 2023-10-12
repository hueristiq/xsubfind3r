package sources

// Source is an interface inherited by each source.
type Source interface {
	// Run takes in configuration which includes keys/tokens and other stuff,
	// and domain as arguments.
	Run(config *Configuration, domain string) <-chan Result
	// Name returns the name of the source.
	Name() string
}

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
