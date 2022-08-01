package runner

import (
	"strings"

	"github.com/hueristiq/subfind3r/internal/configuration"
	"github.com/hueristiq/subfind3r/pkg/agent"
	"github.com/hueristiq/subfind3r/pkg/sources"
)

type Runner struct {
	Options *configuration.Options
	Agent   *agent.Agent
}

func New(options *configuration.Options) *Runner {
	var uses, exclusions []string

	if options.SourcesUse != "" {
		uses = append(uses, strings.Split(options.SourcesUse, ",")...)
	} else {
		uses = append(uses, sources.All...)
	}

	if options.SourcesExclude != "" {
		exclusions = append(exclusions, strings.Split(options.SourcesExclude, ",")...)
	}

	return &Runner{
		Options: options,
		Agent:   agent.New(uses, exclusions),
	}
}

func (runner *Runner) Run() (chan sources.Subdomain, error) {
	subdomains := make(chan sources.Subdomain)

	uniqueMap := make(map[string]sources.Subdomain)
	sourceMap := make(map[string]map[string]struct{})

	keys := runner.Options.YAMLConfig.GetKeys()
	results := runner.Agent.Run(runner.Options.Domain, &keys)

	go func() {
		defer close(subdomains)

		for result := range results {
			if !strings.HasSuffix(result.Value, "."+runner.Options.Domain) {
				continue
			}

			sub := strings.ToLower(result.Value)

			// remove wildcards (`*`)
			sub = strings.ReplaceAll(sub, "*.", "")

			if _, ok := uniqueMap[sub]; !ok {
				sourceMap[sub] = make(map[string]struct{})
			}

			sourceMap[sub][result.Source] = struct{}{}

			if _, ok := uniqueMap[sub]; ok {
				continue
			}

			subdomain := sources.Subdomain{Value: sub, Source: result.Source}

			uniqueMap[sub] = subdomain

			subdomains <- subdomain
		}
	}()

	return subdomains, nil
}
