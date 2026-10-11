// Package input handles user-provided input: stdin detection and domain
// validation.
package input

import (
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// domainPattern accepts a bare hostname: one or more labels separated by
// dots, ending in an alphabetic TLD of at least two characters.
var domainPattern = regexp.MustCompile(`^(?:[a-zA-Z0-9-]+\.)+[a-zA-Z]{2,}$`)

// ValidateDomain reports whether domain is a well-formed hostname.
//
// Parameters:
//   - domain (string): the candidate domain; surrounding whitespace is
//     trimmed before validation.
//
// Returns:
//   - err (error): non-nil when domain is empty or not a valid hostname.
func ValidateDomain(domain string) (err error) {
	domain = strings.TrimSpace(domain)

	if domain == "" {
		err = errors.New("empty domain")

		return
	}

	if !domainPattern.MatchString(domain) {
		err = fmt.Errorf("invalid domain %q", domain)

		return
	}

	return
}

// HasStdin reports whether stdin is redirected from a pipe or file, or is
// a named pipe (FIFO).
func HasStdin() (has bool) {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return
	}

	isPipe := (stat.Mode() & os.ModeCharDevice) == 0
	isFIFO := (stat.Mode() & os.ModeNamedPipe) != 0

	has = isPipe || isFIFO

	return
}
