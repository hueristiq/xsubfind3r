package extractor

import (
	"regexp"
	"sync"
)

var mutex = &sync.Mutex{}

func New(domain string) (extractor *regexp.Regexp, err error) {
	mutex.Lock()
	defer mutex.Unlock()

	pattern := `(\w+[.])*` + regexp.QuoteMeta(domain) + `$`

	extractor, err = regexp.Compile(pattern)
	if err != nil {
		return
	}

	return
}
