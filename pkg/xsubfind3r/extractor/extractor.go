package extractor

import (
	"regexp"
	"sync"
)

var mutex = &sync.Mutex{}

func New(domain string) (extractor *regexp.Regexp, err error) {
	mutex.Lock()
	defer mutex.Unlock()

	pattern := `(?i)[a-zA-Z0-9\*_.-]+\.` + domain

	extractor, err = regexp.Compile(pattern)
	if err != nil {
		return
	}

	return
}
