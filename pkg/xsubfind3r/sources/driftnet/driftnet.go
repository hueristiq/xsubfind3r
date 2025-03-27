// Package driftnet provides an implementation of the sources.Source interface
// for interacting with the Driftnet API.
//
// The Driftnet API aggregates multi-dimensional observation data, including host and certificate
// information, from which subdomains can be extracted. This package defines a Source type that
// implements the Run and Name methods as specified by the sources.Source interface. The Run method
// sends a query to the Driftnet API, processes the JSON response, extracts subdomains from observations
// (from host data and subject certificate data), and streams discovered subdomains or errors via a channel.
package driftnet

import (
	"encoding/json"
	"strings"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
	hqgohttp "go.source.hueristiq.com/http"
	"go.source.hueristiq.com/http/header"
)

// getResultsResponse represents the structure of the JSON response returned by the Driftnet API.
//
// It contains two main sections:
//   - Enrichments: Contains various enrichment data (e.g. CVE, JA4XHash, nuclei-name, product-tag, etc.).
//   - Observations: Contains observed data such as ASN, host, HTTP headers, IP, and subject certificates.
//     The Observations.Host.Values field and Observations.SubjectCert.Values field are used to extract subdomains.
type getResultsResponse struct {
	Enrichments  Enrichments  `json:"enrichments"`
	Observations Observations `json:"observations"`
}

type Enrichments struct {
	CVE           RedactionReason `json:"cve"`
	JA4XHash      RedactionReason `json:"ja4x-hash"`
	JarmFuzzyHash struct{}        `json:"jarm-fuzzyhash"`
	NucleiName    RedactionReason `json:"nuclei-name"`
	ObjMurmur3    struct{}        `json:"obj-murmur3"`
	ProductTag    ProductTag      `json:"product-tag"`
}

type RedactionReason struct {
	RedactionReason string `json:"redaction_reason"`
}

type ProductTag struct {
	Cardinality Cardinality      `json:"cardinality"`
	Other       Cardinality      `json:"other"`
	Values      ProductTagValues `json:"values"`
}

type Cardinality struct {
	Domains   int `json:"domains"`
	Protocols int `json:"protocols"`
}

type ProductTagValues struct {
	Domains   map[string]int `json:"domains"`
	Protocols map[string]int `json:"protocols"`
}

type Observations struct {
	ASN          ObservationData `json:"asn"`
	Entity       ObservationData `json:"entity"`
	GeoCountry   ObservationData `json:"geo-country"`
	Host         HostData        `json:"host"`
	HTTPHeader   HTTPHeaderData  `json:"http-header"`
	IP           IPData          `json:"ip"`
	PortTCP      PortData        `json:"port-tcp"`
	PortUDP      struct{}        `json:"port-udp"`
	ServerBanner ServerBanner    `json:"server-banner"`
	SubjectCert  SubjectCert     `json:"subject;cert"` //nolint: staticcheck
	TitleHTML    TitleHTML       `json:"title;html"`   //nolint: staticcheck
}

type ObservationData struct {
	Cardinality Cardinality               `json:"cardinality"`
	Values      map[string]map[string]int `json:"values"`
}

type HostData struct {
	Cardinality Cardinality               `json:"cardinality"`
	Other       map[string]int            `json:"other"`
	Values      map[string]map[string]int `json:"values"`
}

type HTTPHeaderData struct {
	Cardinality Cardinality                       `json:"cardinality"`
	Other       map[string]int                    `json:"other"`
	Values      map[string]map[string]interface{} `json:"values"`
}

type IPData struct {
	Cardinality Cardinality                       `json:"cardinality"`
	Other       map[string]int                    `json:"other"`
	Values      map[string]map[string]interface{} `json:"values"`
}

type PortData struct {
	Cardinality Cardinality               `json:"cardinality"`
	Values      map[string]map[string]int `json:"values"`
}

type ServerBanner struct {
	Cardinality Cardinality               `json:"cardinality"`
	Values      map[string]map[string]int `json:"values"`
}

type SubjectCert struct {
	Cardinality Cardinality               `json:"cardinality"`
	Values      map[string]map[string]int `json:"values"`
}

type TitleHTML struct {
	Cardinality Cardinality               `json:"cardinality"`
	Values      map[string]map[string]int `json:"values"`
}

// Source represents the Driftnet data source implementation.
// It implements the sources.Source interface, providing functionality
// for retrieving subdomains from the Driftnet API.
type Source struct{}

// Run initiates the process of retrieving subdomain information from the Driftnet API for a given domain.
//
// Parameters:
//   - domain (string): The target domain for which to retrieve subdomains.
//   - cfg (*sources.Configuration): The configuration instance containing API keys,
//     the URL validation function, and any additional settings required by the source.
//
// Returns:
//   - (<-chan sources.Result): A channel that asynchronously emits sources.Result values.
//     Each result is either a discovered subdomain (ResultSubdomain) or an error (ResultError)
//     encountered during the operation.
func (source *Source) Run(domain string, cfg *sources.Configuration) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		getResultsReqURL := "https://api.driftnet.io/v1/multi/summary"
		getResultsReqCFG := &hqgohttp.RequestConfiguration{
			Headers: map[string]string{
				header.Authorization.String(): "Bearer anon",
			},
			Params: map[string]string{
				"summary_limit": "10",
				"timeout":       "30",
				"from":          "2024-12-01",
				"to":            "2024-12-11",
				"field":         "host:" + domain,
			},
		}

		getResultsRes, err := hqgohttp.Get(getResultsReqURL, getResultsReqCFG)
		if err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			return
		}

		var getResultsResData getResultsResponse

		if err = json.NewDecoder(getResultsRes.Body).Decode(&getResultsResData); err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: source.Name(),
				Error:  err,
			}

			results <- result

			getResultsRes.Body.Close()

			return
		}

		getResultsRes.Body.Close()

		for _, v := range getResultsResData.Observations.Host.Values {
			for subdomain := range v {
				if subdomain != domain && !strings.HasSuffix(subdomain, "."+domain) {
					continue
				}

				result := sources.Result{
					Type:   sources.ResultSubdomain,
					Source: source.Name(),
					Value:  subdomain,
				}

				results <- result
			}
		}

		for _, value := range getResultsResData.Observations.SubjectCert.Values {
			for entry := range value {
				subdomains := cfg.Extractor.FindAllString(entry, -1)

				for i := range subdomains {
					subdomain := subdomains[i]

					if subdomain != domain && !strings.HasSuffix(subdomain, "."+domain) {
						continue
					}

					result := sources.Result{
						Type:   sources.ResultSubdomain,
						Source: source.Name(),
						Value:  subdomain,
					}

					results <- result
				}
			}
		}
	}()

	return results
}

// Name returns the unique identifier for the data source.
// This identifier is used for logging, debugging, and associating results with the correct data source.
//
// Returns:
//   - name (string): The unique identifier for the data source.
func (source *Source) Name() (name string) {
	return sources.DRIFTNET
}
