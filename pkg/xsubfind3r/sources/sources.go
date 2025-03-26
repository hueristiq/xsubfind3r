// Package sources provides the core interfaces, types, and constants required for integrating
// multiple data sources into the application.
//
// This package defines the Source interface which every data source implementation must satisfy.
// It standardizes the functionality for subdomain enumeration, ensuring consistent behavior across
// various integrations. In addition, the package provides configuration types for managing API keys,
// regular expression extractors, and other settings needed for interacting with external data sources.
// The Result and ResultType types are used to encapsulate the outcomes of data collection operations,
// making it easy to report successful subdomain discoveries or errors.
//
// Supported data sources are defined by a set of constants (e.g., ANUBIS, SHODAN, GITHUB, etc.) and a
// List slice that can be used to iterate over or validate available integrations.
package sources

import (
	"crypto/rand"
	"errors"
	"fmt"
	"math/big"
	"regexp"
)

// Source is the interface that every data source implementation must satisfy.
// It standardizes the required functionality, ensuring uniform behavior across
// different integrations. Implementers of Source must define two methods:
//
//   - Run: Initiates data collection for a given domain using the provided configuration,
//     returning results asynchronously through a channel.
//   - Name: Returns the unique identifier (name) of the data source for logging and reporting.
type Source interface {
	// Run initiates the data collection or scanning process for a specified domain.
	// The method accepts a domain name and a pointer to a Configuration instance,
	// and returns a read-only channel through which results (of type Result) are streamed.
	//
	// Parameters:
	//   - domain (string): A string representing the target domain for data collection.
	//   - cfg (*Configuration): A pointer to a Configuration struct containing API keys, regular expressions,
	//          and any other settings needed for interacting with the data source.
	//
	// Returns:
	//   - (<-chan Result): A read-only channel that asynchronously emits Result values,
	//     allowing the caller to process subdomain data or errors as they become available.
	Run(domain string, cfg *Configuration) <-chan Result

	// Name returns the unique name of the data source.
	//
	// This identifier is used for distinguishing among multiple data sources,
	// especially when logging activity or compiling results from several integrations.
	//
	// Returns:
	//   - name (string): A string that uniquely identifies the data source.
	Name() (name string)
}

// Configuration holds the settings and parameters used by the Finder and its sources.
//
// Fields:
//   - Keys (Keys): Contains the API keys for various data sources, allowing authenticated access.
//   - Extractor (*regexp.Regexp): A compiled regular expression used to extract or validate
//     domain-related patterns, ensuring consistent parsing and validation.
type Configuration struct {
	Keys      Keys
	Extractor *regexp.Regexp
}

// Keys stores API keys for different data sources. Each field represents a collection of API keys
// for a specific source, and is defined using the SourceKeys type (a slice of strings). These keys are
// used for authentication when interacting with external APIs or services.
//
// Fields (each field is a SourceKeys slice):
//   - Bevigil: API keys for the Bevigil data source.
//   - BuiltWith: API keys for the BuiltWith data source.
//   - Censys: API keys for the Censys data source.
//   - Certspotter: API keys for the Certspotter data source.
//   - Chaos: API keys for the Chaos data source.
//   - Fullhunt: API keys for the Fullhunt data source.
//   - GitHub: API keys for the GitHub data source.
//   - Intelx: API keys for the Intelx data source.
//   - LeakIX: API keys for the LeakIX data source.
//   - SecurityTrails: API keys for the SecurityTrails data source.
//   - Shodan: API keys for the Shodan data source.
//   - URLScan: API keys for the URLScan data source.
//   - VirusTotal: API keys for the VirusTotal data source.
type Keys struct {
	Bevigil        SourceKeys `yaml:"bevigil"`
	BuiltWith      SourceKeys `yaml:"builtwith"`
	Censys         SourceKeys `yaml:"censys"`
	Certspotter    SourceKeys `yaml:"certspotter"`
	Chaos          SourceKeys `yaml:"chaos"`
	Fullhunt       SourceKeys `yaml:"fullhunt"`
	GitHub         SourceKeys `yaml:"github"`
	Intelx         SourceKeys `yaml:"intelx"`
	LeakIX         SourceKeys `yaml:"leakix"`
	SecurityTrails SourceKeys `yaml:"securitytrails"`
	Shodan         SourceKeys `yaml:"shodan"`
	URLScan        SourceKeys `yaml:"urlscan"`
	VirusTotal     SourceKeys `yaml:"virustotal"`
}

// SourceKeys is a slice of strings where each element represents an API key for a specific source.
// This structure supports maintaining multiple keys for a single source, which is useful for key
// rotation or providing fallback options if one key becomes invalid.
type SourceKeys []string

// PickRandom selects and returns a random API key from the SourceKeys slice.
// It uses a cryptographically secure random number generator to ensure a secure and unbiased selection.
// If the slice is empty, it returns the ErrNoKeys error.
//
// Returns:
//   - key (string): A randomly selected API key from the slice.
//   - err (error): An error if the slice is empty or if randomness generation fails.
func (k SourceKeys) PickRandom() (key string, err error) {
	length := len(k)

	if length == 0 {
		err = ErrNoKeys

		return
	}

	maximum := big.NewInt(int64(length))

	var indexBig *big.Int

	indexBig, err = rand.Int(rand.Reader, maximum)
	if err != nil {
		err = fmt.Errorf("failed to generate random index: %w", err)

		return
	}

	index := indexBig.Int64()

	key = k[index]

	return
}

// Result represents the outcome of subdomain discovery.
// It encapsulates details about the result, including its type, the originating source,
// the actual data (if available), and any error encountered during the operation.
//
// Fields:
//   - Type (ResultType): Specifies the kind of result (e.g., subdomain or error).
//   - Source (string): Identifies the source that produced this result (e.g., "crtsh", "shodan").
//   - Value (string): Contains the actual subdomain retrieved from the source.
//     This field is empty if the result is an error.
//   - Error (error): Holds the error encountered during the operation, if any. If no error
//     occurred, this field is nil.
type Result struct {
	Type   ResultType
	Source string
	Value  string
	Error  error
}

// ResultType defines the category of a Result using an integer enumeration.
// It allows for distinguishing between different types of outcomes produced by sources.
//
// Enumeration Values:
//   - ResultSubdomain: Indicates a successful result containing a subdomain or related data.
//   - ResultError: Represents an outcome where an error occurred during the data collection.
type ResultType int

// Constants representing the types of results that can be produced by a data source.
//
// List of Constants:
//   - ResultSubdomain: Represents a successful result containing subdomain data or
//     other relevant information.
//   - ResultError: Indicates an error encountered during the operation, with details
//     provided in the Error field of the Result.
const (
	ResultSubdomain ResultType = iota
	ResultError
)

// Supported data source constants.
//
// The following constants define the names of supported data sources.
// Each constant is used as a unique identifier for its corresponding data source,
// which might provide different types of data (e.g., subdomains, SSL/TLS certificates,
// historical records, vulnerability data, etc.).
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
	OPENTHREATEXCHANGE = "otx"
	SECURITYTRAILS     = "securitytrails"
	SHODAN             = "shodan"
	SUBDOMAINCENTER    = "subdomaincenter"
	URLSCAN            = "urlscan"
	WAYBACK            = "wayback"
	VIRUSTOTAL         = "virustotal"
)

// ErrNoKeys is a sentinel error returned when a SourceKeys slice contains no API keys.
// This error is used to signal that an operation requiring an API key cannot proceed
// because no keys are available.
var ErrNoKeys = errors.New("no keys available for the source")

// List is a collection of all supported source names.
//
// This slice provides a convenient way to iterate over, validate, or dynamically configure
// the data sources available in the application. Developers can use List to:.
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
	OPENTHREATEXCHANGE,
	SECURITYTRAILS,
	SHODAN,
	SUBDOMAINCENTER,
	URLSCAN,
	WAYBACK,
	VIRUSTOTAL,
}
