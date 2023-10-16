package scraper

import "github.com/hueristiq/xsubfind3r/pkg/scraper/sources"

type Options struct {
	SourcesToUSe     []string
	SourcesToExclude []string
	Keys             sources.Keys
}
