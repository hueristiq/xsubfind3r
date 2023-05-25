package sources

type Source interface {
	Run(config *Configuration) (subdomains chan Subdomain)
	Name() string
}
