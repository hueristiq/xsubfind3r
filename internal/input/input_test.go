package input_test

import (
	"testing"

	"github.com/hueristiq/xsubfind3r/internal/input"
)

func TestValidateDomain(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		domain  string
		wantErr bool
	}{
		{
			name:   "plain domain",
			domain: "example.com",
		},
		{
			name:   "subdomain",
			domain: "a.b.example.com",
		},
		{
			name:   "hyphenated label",
			domain: "x-sub.example.com",
		},
		{
			name:   "surrounding whitespace is trimmed",
			domain: "  example.com  ",
		},
		{
			name:    "empty",
			domain:  "",
			wantErr: true,
		},
		{
			name:    "whitespace only",
			domain:  "   ",
			wantErr: true,
		},
		{
			name:    "path traversal",
			domain:  "../../tmp/x",
			wantErr: true,
		},
		{
			name:    "path separator",
			domain:  "a/example.com",
			wantErr: true,
		},
		{
			name:    "single label",
			domain:  "localhost",
			wantErr: true,
		},
		{
			name:    "numeric tld",
			domain:  "example.123",
			wantErr: true,
		},
		{
			name:    "url with scheme",
			domain:  "https://example.com",
			wantErr: true,
		},
		{
			name:    "wildcard",
			domain:  "*.example.com",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			err := input.ValidateDomain(tt.domain)

			if (err != nil) != tt.wantErr {
				t.Fatalf("ValidateDomain(%q) error = %v, wantErr %v", tt.domain, err, tt.wantErr)
			}
		})
	}
}
