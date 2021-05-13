package sources

import (
	"regexp"
	"sync"
)

var extractorMutex = &sync.Mutex{}

func NewExtractor(domain string) (*regexp.Regexp, error) {
	extractorMutex.Lock()
	defer extractorMutex.Unlock()

	extractor, err := regexp.Compile(`[a-zA-Z0-9\*_.-]+\.` + domain)
	if err != nil {
		return nil, err
	}

	return extractor, nil
}
