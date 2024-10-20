package sources

// Source is an interface that defines methods for a data source.
// Any source that implements this interface should define a process to run
// data collection or scanning based on a configuration and domain,
// and provide a way to retrieve the source name.
type Source interface {
	// Run starts the data collection or scanning process for a specific domain.
	// It takes in a Configuration and a domain string as input and returns a channel
	// of Result structs, which will asynchronously emit results from the data source.
	// The use of channels allows for concurrent processing and retrieval of data.
	Run(cfg *Configuration, domain string) <-chan Result

	// Name returns the name of the source. This can be used to identify the data source
	// implementing the interface. Useful for logging, reporting, or debugging purposes.
	Name() string
}

// Constants representing the names of different data sources.
// These constants can be used to refer to various OSINT (Open-Source Intelligence)
// sources, threat intelligence platforms, or search engines that are commonly used
// for gathering information about domains, IP addresses, or other targets.
const (
	ANUBIS             = "anubis"             // Anubis is an OSINT tool for gathering domain information.
	BEVIGIL            = "bevigil"            // Bevigil is an OSINT platform focused on vulnerabilities in mobile apps.
	BUILTWITH          = "builtwith"          // BuiltWith is a service for analyzing website technologies.
	CENSYS             = "censys"             // Censys is a search engine for internet-connected devices and their data.
	CERTIFICATEDETAILS = "certificatedetails" // CertificateDetails provides SSL/TLS certificate information.
	CERTSPOTTER        = "certspotter"        // CertSpotter monitors SSL/TLS certificates for domains.
	CHAOS              = "chaos"              // Chaos by ProjectDiscovery is a source for subdomain enumeration.
	COMMONCRAWL        = "commoncrawl"        // Common Crawl is an open repository of web data.
	CRTSH              = "crtsh"              // crt.sh is a certificate transparency log search engine.
	FULLHUNT           = "fullhunt"           // FullHunt is a platform for attack surface monitoring.
	GITHUB             = "github"             // GitHub is a source for finding code repositories and related metadata.
	HACKERTARGET       = "hackertarget"       // HackerTarget provides security scanning services.
	INTELLIGENCEX      = "intelx"             // Intelligence X is a search engine for intelligence gathering.
	LEAKIX             = "leakix"             // LeakIX is a search engine for finding leaked and exposed data.
	OPENTHREATEXCHANGE = "otx"                // Open Threat Exchange (OTX) is a collaborative threat intelligence platform.
	SECURITYTRAILS     = "securitytrails"     // SecurityTrails offers a comprehensive API for domain information.
	SHODAN             = "shodan"             // Shodan is a search engine for internet-connected devices and vulnerabilities.
	SUBDOMAINCENTER    = "subdomaincenter"    // SubdomainCenter is a tool for subdomain enumeration.
	URLSCAN            = "urlscan"            // URLScan.io is a service for scanning websites and collecting URLs.
	WAYBACK            = "wayback"            // Wayback Machine is an internet archive for historical website snapshots.
	VIRUSTOTAL         = "virustotal"         // VirusTotal is a platform for scanning files and URLs for malware.
)

// List contains a collection of all available source names.
// This array is useful for iterating over or referencing the supported data sources
// in loops, validation logic, or dynamic configurations.
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
