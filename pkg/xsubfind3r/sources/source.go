package sources

type Source interface {
	Run(config *Configuration, domain string) (subdomains chan Subdomain)
	Name() string
}
