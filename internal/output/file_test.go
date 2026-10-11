package output_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/hueristiq/xsubfind3r/internal/output"
	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

const (
	subdomainValue = "api.example.com"
	subdomainLine  = subdomainValue + "\n"
	sourceName     = "crtsh"
)

func TestWriterTXT(t *testing.T) {
	t.Parallel()

	writer := output.NewWriter()

	buffer := &bytes.Buffer{}

	result := sources.Result{
		Type:   sources.ResultSubdomain,
		Source: sourceName,
		Value:  subdomainValue,
	}

	if err := writer.Write(buffer, "example.com", result); err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if buffer.Len() != 0 {
		t.Fatalf("buffer has %q before flush, want empty", buffer.String())
	}

	if err := writer.Flush(buffer); err != nil {
		t.Fatalf("Flush() error: %v", err)
	}

	if got, want := buffer.String(), subdomainLine; got != want {
		t.Fatalf("output = %q, want %q", got, want)
	}
}

func TestWriterJSONL(t *testing.T) {
	t.Parallel()

	writer := output.NewWriter()
	writer.SetFormatToJSONL()

	buffer := &bytes.Buffer{}

	result := sources.Result{
		Type:   sources.ResultSubdomain,
		Source: sourceName,
		Value:  subdomainValue,
	}

	if err := writer.Write(buffer, "example.com", result); err != nil {
		t.Fatalf("Write() error: %v", err)
	}

	if err := writer.Flush(buffer); err != nil {
		t.Fatalf("Flush() error: %v", err)
	}

	line := strings.TrimSuffix(buffer.String(), "\n")

	want := `{"domain":"example.com","subdomain":"api.example.com","source":"crtsh"}`

	if line != want {
		t.Fatalf("output = %q, want %q", line, want)
	}
}

func TestWriterIndependentStreams(t *testing.T) {
	t.Parallel()

	writer := output.NewWriter()

	first := &bytes.Buffer{}
	second := &bytes.Buffer{}

	result := sources.Result{
		Type:   sources.ResultSubdomain,
		Source: sourceName,
		Value:  subdomainValue,
	}

	for _, buffer := range []*bytes.Buffer{first, second} {
		if err := writer.Write(buffer, "example.com", result); err != nil {
			t.Fatalf("Write() error: %v", err)
		}
	}

	if err := writer.Flush(first); err != nil {
		t.Fatalf("Flush() error: %v", err)
	}

	if second.Len() != 0 {
		t.Fatalf("second buffer has %q, want empty until flushed", second.String())
	}

	if err := writer.Flush(second); err != nil {
		t.Fatalf("Flush() error: %v", err)
	}

	if got, want := first.String(), subdomainLine; got != want {
		t.Fatalf("first output = %q, want %q", got, want)
	}

	if got, want := second.String(), subdomainLine; got != want {
		t.Fatalf("second output = %q, want %q", got, want)
	}
}
