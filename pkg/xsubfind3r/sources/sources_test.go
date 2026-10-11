package sources_test

import (
	"reflect"
	"testing"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

const nestedDomain = "a.b.example.com"

func TestConfigurationExtract(t *testing.T) {
	t.Parallel()

	cfg := &sources.Configuration{
		Domain:    "example.com",
		Extractor: sources.CompileExtractor("example.com"),
	}

	tests := []struct {
		name  string
		input string
		want  []string
	}{
		{
			name:  "plain subdomain",
			input: "sub.example.com",
			want:  []string{"sub.example.com"},
		},
		{
			name:  "nested subdomain",
			input: "a.b.example.com",
			want:  []string{"a.b.example.com"},
		},
		{
			name:  "multiple subdomains",
			input: "foo.bar.example.com baz.example.com",
			want:  []string{"foo.bar.example.com", "baz.example.com"},
		},
		{
			name:  "comma separated",
			input: "a.example.com,b.example.com",
			want:  []string{"a.example.com", "b.example.com"},
		},
		{
			name:  "with port",
			input: "x.example.com:443",
			want:  []string{"x.example.com"},
		},
		{
			name:  "uppercase input",
			input: "SUB.EXAMPLE.COM",
			want:  []string{"SUB.EXAMPLE.COM"},
		},
		{
			name:  "hyphenated labels",
			input: "x-sub.example.com",
			want:  []string{"x-sub.example.com"},
		},
		{
			name:  "lookalike prefix",
			input: "notexample.com",
			want:  nil,
		},
		{
			name:  "lookalike suffix",
			input: "example.com.evil.com",
			want:  nil,
		},
		{
			name:  "lookalike with subdomain",
			input: "sub.example.com.evil.com",
			want:  nil,
		},
		{
			name:  "wildcard prefix",
			input: "*.example.com",
			want:  nil,
		},
		{
			name:  "empty label",
			input: "sub..example.com",
			want:  nil,
		},
		{
			name:  "apex alone is not extracted",
			input: "example.com",
			want:  nil,
		},
		{
			name:  "trailing letter",
			input: "sub.example.comm",
			want:  nil,
		},
		{
			name:  "no separator between hosts yields the longest host",
			input: "a.example.comb.example.com",
			want:  []string{"a.example.comb.example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got := cfg.Extract(tt.input)

			if !reflect.DeepEqual(got, tt.want) {
				t.Fatalf("Extract(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestConfigurationExtractEscapedDomain(t *testing.T) {
	t.Parallel()

	cfg := &sources.Configuration{
		Domain:    nestedDomain,
		Extractor: sources.CompileExtractor(nestedDomain),
	}

	got := cfg.Extract("x.a.b.example.com")

	want := []string{"x.a.b.example.com"}

	if !reflect.DeepEqual(got, want) {
		t.Fatalf("Extract() = %v, want %v", got, want)
	}
}
