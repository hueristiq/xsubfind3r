package driftnet

import (
	"encoding/json"
	"fmt"
	"strings"

	hqgohttp "github.com/hueristiq/hq-go-http"
	hqgohttpheader "github.com/hueristiq/hq-go-http/header"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

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

type Source struct{}

func (s *Source) Name() (name string) {
	name = sources.DRIFTNET

	return
}

func (s *Source) Run(cfg *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		getResultsReqURL := "https://api.driftnet.io/v1/multi/summary"
		getResultsReqCFG := &hqgohttp.RequestConfiguration{
			Headers: []hqgohttp.Header{
				hqgohttp.NewSetHeader(hqgohttpheader.Authorization.String(), "Bearer anon"),
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
				Source: s.Name(),
				Error:  fmt.Errorf("request failed: %w", err),
			}

			results <- result

			return
		}

		var getResultsResData getResultsResponse

		if err = json.NewDecoder(getResultsRes.Body).Decode(&getResultsResData); err != nil {
			result := sources.Result{
				Type:   sources.ResultError,
				Source: s.Name(),
				Error:  fmt.Errorf("failed to parse JSON response: %w", err),
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
					Source: s.Name(),
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
						Source: s.Name(),
						Value:  subdomain,
					}

					results <- result
				}
			}
		}
	}()

	return results
}

func (s *Source) UseKeys(keys ...string) {
}

var _ sources.Source = (*Source)(nil)

func New() (source sources.Source) {
	source = &Source{}

	return
}
