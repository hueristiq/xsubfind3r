package scraper

import "github.com/hueristiq/xsubfind3r/pkg/scraper/sources"

type Options struct {
	SourcesToExclude []string
	SourcesToUSe     []string
	Keys             sources.Keys
}
