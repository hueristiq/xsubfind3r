package driftnet

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	hqgohttpheader "github.com/hueristiq/hq-lib-http-go/header"

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
	SubjectCert  SubjectCert     `json:"subject;cert"`
	TitleHTML    TitleHTML       `json:"title;html"`
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
	name = sources.Driftnet

	return
}

func (s *Source) NeedsKeys() (needs bool) {
	needs = false

	return
}

func (s *Source) UseKeys(keys ...string) {
}

func (s *Source) Run(ctx context.Context, cfg *sources.Configuration, domain string) <-chan sources.Result {
	results := make(chan sources.Result)

	go func() {
		defer close(results)

		getResultsReqURL := "https://api.driftnet.io/v1/multi/summary"

		// The public endpoint accepts the documented anonymous bearer
		// token; a rolling 30-day window keeps the query current instead
		// of pinned to a fixed period.
		to := time.Now().UTC()
		from := to.AddDate(0, 0, -30)

		getResultsRes, err := cfg.Client.NewRequest().
			URL(getResultsReqURL).
			Param("summary_limit", "10").
			Param("timeout", "30").
			Param("from", from.Format("2006-01-02")).
			Param("to", to.Format("2006-01-02")).
			Param("field", "host:"+domain).
			SetHeader(hqgohttpheader.Authorization.String(), "Bearer anon").
			Do(ctx)
		if err != nil {
			result := sources.ErrorResult(s.Name(), fmt.Errorf("request failed: %w", err))

			results <- result

			return
		}

		if err = sources.CheckResponse(getResultsRes); err != nil {
			result := sources.ErrorResult(s.Name(), err)

			results <- result

			_ = getResultsRes.Body.Close()

			return
		}

		var getResultsResData getResultsResponse

		if err = json.NewDecoder(getResultsRes.Body).Decode(&getResultsResData); err != nil {
			result := sources.ErrorResult(s.Name(), fmt.Errorf("failed to parse JSON response: %w", err))

			results <- result

			_ = getResultsRes.Body.Close()

			return
		}

		_ = getResultsRes.Body.Close()

		for _, v := range getResultsResData.Observations.Host.Values {
			for subdomain := range v {
				subdomain, valid := cfg.Validator(subdomain)
				if !valid {
					continue
				}

				result := sources.SubdomainResult(s.Name(), subdomain)

				results <- result
			}
		}

		for _, value := range getResultsResData.Observations.SubjectCert.Values {
			for entry := range value {
				for _, candidate := range sources.ExtractSubdomains(cfg.Extractor, entry) {
					subdomain, valid := cfg.Validator(candidate)
					if !valid {
						continue
					}

					result := sources.SubdomainResult(s.Name(), subdomain)

					results <- result
				}
			}
		}
	}()

	return results
}

var _ sources.Source = (*Source)(nil)

func New() (source sources.Source) {
	source = &Source{}

	return
}
