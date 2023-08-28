package xsubfind3r

import "github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"

type Options struct {
	SourcesToExclude []string
	SourcesToUSe     []string
	Keys             sources.Keys
}
