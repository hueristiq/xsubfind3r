package extractor

import (
	"regexp"
	"sync"
)

var mutex = &sync.Mutex{}

func New(domain string) (extractor *regexp.Regexp, err error) {
	mutex.Lock()
	defer mutex.Unlock()

	extractor, err = regexp.Compile(`(?i)[a-zA-Z0-9\*_.-]+\.` + domain)
	if err != nil {
		return
	}

	return
}
