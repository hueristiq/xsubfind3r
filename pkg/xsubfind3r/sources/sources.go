package sources

// Source is an interface that defines the blueprint for a data source.
// Any data source integrated into the application must implement this interface.
// It provides methods for initiating data collection and identifying the source.
//
// Methods:
//   - Run: Executes the data collection or scanning process for a specific domain
//     and configuration, returning results asynchronously through a channel.
//   - Name: Returns the name of the data source, which is helpful for logging,
//     debugging, or identifying the source in reports.
type Source interface {
	// Run starts the data collection process for the specified domain using the provided
	// configuration. It returns a channel that emits `Result` structs asynchronously,
	// allowing concurrent processing and streaming of results.
	//
	// Parameters:
	// - cfg *Configuration: The configuration settings, such as API keys and options,
	//   required to interact with the data source.
	// - domain string: The target domain for which data is to be collected.
	//
	// Returns:
	// - <-chan Result: A channel that streams the results produced by the data source.
	Run(cfg *Configuration, domain string) <-chan Result
	// Name retrieves the unique identifier or name of the data source. This is primarily
	// used for distinguishing between multiple sources, logging activity, and reporting
	// during subdomain enumeration.
	//
	// Returns:
	// - string: The name of the data source.
	Name() string
}

// The following constants define the names of supported data sources.
// These constants are used as unique identifiers for the corresponding sources.
// Each source provides different types of data, such as subdomains, SSL/TLS certificates,
// historical records, or vulnerability data.
//
// List of Constants:
//
// - ANUBIS: An OSINT tool for gathering domain information.
// - BEVIGIL: Focuses on vulnerabilities in mobile applications.
// - BUILTWITH: Analyzes website technologies and services.
// - CENSYS: Search engine for internet-connected devices and related metadata.
// - CERTIFICATEDETAILS: Fetches information about SSL/TLS certificates.
// - CERTSPOTTER: Monitors certificate transparency logs.
// - CHAOS: ProjectDiscovery’s service for subdomain enumeration.
// - COMMONCRAWL: Repository of open web data.
// - CRTSH: Certificate transparency log search engine.
// - FULLHUNT: Platform for attack surface monitoring.
// - GITHUB: Searches code repositories for relevant data.
// - HACKERTARGET: Offers security scanning and OSINT capabilities.
// - INTELLIGENCEX: Search engine for intelligence data.
// - LEAKIX: Search engine for leaked and exposed data.
// - OPENTHREATEXCHANGE: Collaborative threat intelligence platform.
// - SECURITYTRAILS: Provides comprehensive domain and DNS information.
// - SHODAN: Search engine for internet-connected devices and vulnerabilities.
// - SUBDOMAINCENTER: Tool dedicated to subdomain enumeration.
// - URLSCAN: Service for website scanning and URL collection.
// - WAYBACK: Internet archive for historical website snapshots.
// - VIRUSTOTAL: Malware scanning for files and URLs.
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
	FULLHUNT           = "fullhunt"
	GITHUB             = "github"
	HACKERTARGET       = "hackertarget"
	INTELLIGENCEX      = "intelx"
	LEAKIX             = "leakix"
	OPENTHREATEXCHANGE = "otx"
	SECURITYTRAILS     = "securitytrails"
	SHODAN             = "shodan"
	SUBDOMAINCENTER    = "subdomaincenter"
	URLSCAN            = "urlscan"
	WAYBACK            = "wayback"
	VIRUSTOTAL         = "virustotal"
)

// List is a collection of all supported source names.
// It provides a convenient way to iterate over, validate, or configure the data sources dynamically.
// Developers can use this list for tasks such as enabling specific sources or verifying
// that a provided source name is valid.
//
// Usage:
// - Iterate over the List to dynamically load supported sources.
// - Validate user input by checking against the entries in the List.
//
// Example:
//
//	for _, source := range List {
//	    fmt.Println("Supported source:", source)
//	}
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
	FULLHUNT,
	GITHUB,
	HACKERTARGET,
	INTELLIGENCEX,
	LEAKIX,
	OPENTHREATEXCHANGE,
	SECURITYTRAILS,
	SHODAN,
	SUBDOMAINCENTER,
	URLSCAN,
	WAYBACK,
	VIRUSTOTAL,
}
