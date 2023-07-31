package sources

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
