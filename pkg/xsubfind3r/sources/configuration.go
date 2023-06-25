package sources

import "regexp"

type Configuration struct {
	Keys               Keys
	Domain             string
	IncludeSubdomains  bool
	ParseWaybackRobots bool
	ParseWaybackSource bool
	SubdomainsRegex    *regexp.Regexp
}

type Keys struct {
	Bevigil  []string `yaml:"bevigil"`
	Censys   []string `yaml:"censys"`
	Chaos    []string `yaml:"chaos"`
	Fullhunt []string `yaml:"fullhunt"`
	GitHub   []string `yaml:"github"`
	Intelx   []string `yaml:"intelx"`
	URLScan  []string `yaml:"urlscan"`
}
