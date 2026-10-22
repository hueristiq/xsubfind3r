package sources_test

import (
	"errors"
	"io"
	"net/http"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

const (
	testSource    = "crtsh"
	testSubdomain = "a.example.com"
)

func TestCheckResponse(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		statusCode  int
		status      string
		body        string
		secrets     []string
		wantErr     bool
		errorSubstr []string
	}{
		{name: "ok", statusCode: http.StatusOK, status: "200 OK", body: "ignored"},
		{name: "no content", statusCode: http.StatusNoContent, status: "204 No Content"},
		{name: "redirect", statusCode: http.StatusMovedPermanently, status: "301 Moved Permanently", body: "moved", wantErr: true, errorSubstr: []string{"301", "moved"}},
		{name: "not found", statusCode: http.StatusNotFound, status: "404 Not Found", body: "  not found  ", wantErr: true, errorSubstr: []string{"unexpected status 404 Not Found: not found"}},
		{name: "redacted", statusCode: http.StatusInternalServerError, status: "500 Internal Server Error", body: "secret-value leaked", secrets: []string{"secret-value"}, wantErr: true, errorSubstr: []string{"[REDACTED]", "unexpected status 500"}},
		{name: "empty body", statusCode: http.StatusBadGateway, status: "502 Bad Gateway", wantErr: true, errorSubstr: []string{"unexpected status 502 Bad Gateway"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			response := &http.Response{
				StatusCode: test.statusCode,
				Status:     test.status,
				Body:       io.NopCloser(strings.NewReader(test.body)),
			}

			err := sources.CheckResponse(response, test.secrets...)

			_ = response.Body.Close()

			if !test.wantErr {
				if err != nil {
					t.Fatalf("CheckResponse: %v", err)
				}

				return
			}

			if err == nil {
				t.Fatal("CheckResponse returned no error")
			}

			for _, substring := range test.errorSubstr {
				if !strings.Contains(err.Error(), substring) {
					t.Fatalf("error = %q, want substring %q", err.Error(), substring)
				}
			}

			if len(test.secrets) > 0 && strings.Contains(err.Error(), test.secrets[0]) {
				t.Fatalf("error leaked secret: %q", err.Error())
			}
		})
	}
}

func TestCheckResponseNil(t *testing.T) {
	t.Parallel()

	if err := sources.CheckResponse(nil); err == nil {
		t.Fatal("CheckResponse(nil) returned no error")
	}
}

func TestResultConstructors(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		result sources.Result
		want   sources.Result
	}{
		{
			name:   "error",
			result: sources.ErrorResult(testSource, errors.New("boom")),
			want:   sources.Result{Type: sources.ResultError, Source: testSource, Error: errors.New("boom")},
		},
		{
			name:   "warning",
			result: sources.WarningResult(testSource, "capped"),
			want:   sources.Result{Type: sources.ResultWarning, Source: testSource, Value: "capped"},
		},
		{
			name:   "subdomain",
			result: sources.SubdomainResult(testSource, testSubdomain),
			want:   sources.Result{Type: sources.ResultSubdomain, Source: testSource, Value: testSubdomain},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			if test.result.Type != test.want.Type || test.result.Source != test.want.Source || test.result.Value != test.want.Value {
				t.Fatalf("result = %+v, want %+v", test.result, test.want)
			}

			if test.want.Error == nil {
				if test.result.Error != nil {
					t.Fatalf("error = %v, want nil", test.result.Error)
				}

				return
			}

			if test.result.Error == nil || test.result.Error.Error() != test.want.Error.Error() {
				t.Fatalf("error = %v, want %v", test.result.Error, test.want.Error)
			}
		})
	}
}

func TestExtractSubdomains(t *testing.T) {
	t.Parallel()

	extractor := regexp.MustCompile(`(?i)(?:^|[^a-z0-9_.-])(((?:[a-z0-9_](?:[a-z0-9_-]*[a-z0-9_])?\.)+)example\.com)`)

	tests := []struct {
		name string
		text string
		want []string
	}{
		{name: "plain", text: testSubdomain, want: []string{testSubdomain}},
		{name: "a label", text: "api-v2.example.com", want: []string{"api-v2.example.com"}},
		{name: "nested", text: "x.y.example.com", want: []string{"x.y.example.com"}},
		{name: "terminal dot eof", text: testSubdomain + ".", want: []string{testSubdomain}},
		{name: "terminal dot whitespace", text: testSubdomain + ". next", want: []string{testSubdomain}},
		{name: "terminal dot comma", text: testSubdomain + ", next", want: []string{testSubdomain}},
		{name: "longer hostname", text: testSubdomain + ".evil.com", want: []string{}},
		{name: "hostname suffix", text: testSubdomain + "-other", want: []string{}},
		{name: "hostname character", text: testSubdomain + "x", want: []string{}},
		{name: "inside text", text: "see (" + testSubdomain + ") here", want: []string{testSubdomain}},
		{name: "encoded slash prefix", text: "%2F" + testSubdomain, want: []string{testSubdomain}},
		{name: "encoded slash prefix lowercase", text: "%2f" + testSubdomain, want: []string{testSubdomain}},
		{name: "encoded slash empty label", text: "%2f.example.com", want: []string{}},
		{name: "legitimate two f label", text: "2f.example.com", want: []string{"2f.example.com"}},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			got := sources.ExtractSubdomains(extractor, test.text)

			if !slices.Equal(got, test.want) {
				t.Fatalf("ExtractSubdomains(%q) = %v, want %v", test.text, got, test.want)
			}
		})
	}
}
