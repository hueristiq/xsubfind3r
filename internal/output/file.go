package output

import (
	"bufio"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/hueristiq/xsubfind3r/pkg/xsubfind3r/sources"
)

// Writer streams enumeration results in plain-text or JSONL format. Each
// destination (stdout, an output file) gets its own buffered stream, so
// results are written in bulk instead of allocating and flushing a buffer
// per result.
type Writer struct {
	format  format
	writers map[io.Writer]*bufio.Writer
}

// SetFormatToJSONL switches the writer to JSONL output.
func (w *Writer) SetFormatToJSONL() {
	w.format = formatJSONL
}

// CreateFile creates (or opens for appending) the output file at path,
// creating the containing directory with 0750 permissions and the file
// with 0600. The file extension is normalized to the active format unless
// it already matches.
//
// Parameters:
//   - path (string): the output file location.
//
// Returns:
//   - file (*os.File): the open file.
//   - err (error): non-nil when path is empty or the file cannot be
//     created.
func (w *Writer) CreateFile(path string) (file *os.File, err error) {
	if path == "" {
		err = ErrNoFilePathSpecified

		return
	}

	extension := filepath.Ext(path)

	switch w.format {
	case formatTXT:
		if extension != ".txt" {
			path += ".txt"
		}
	case formatJSONL:
		if extension != ".json" {
			path += ".json"
		}
	}

	directory := filepath.Dir(path)

	if directory != "" {
		if _, err = os.Stat(directory); os.IsNotExist(err) {
			if err = os.MkdirAll(directory, 0o750); err != nil {
				return
			}
		}
	}

	file, err = os.OpenFile(path, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600) //nolint:gosec // path is the user-specified output file location.
	if err != nil {
		return
	}

	return
}

// Write buffers a result into writer's stream in the active format; use
// [Writer.Flush] to push the buffered data to the destination.
//
// Parameters:
//   - writer (io.Writer): the destination stream.
//   - domain (string): the domain the result belongs to.
//   - result (sources.Result): the result to write.
//
// Returns:
//   - err (error): non-nil when the result cannot be buffered.
func (w *Writer) Write(writer io.Writer, domain string, result sources.Result) (err error) {
	bw := w.bufferedWriter(writer)

	switch w.format {
	case formatTXT:
		if _, err = bw.WriteString(result.Value + "\n"); err != nil {
			return
		}
	case formatJSONL:
		data := resultForJSONL{
			Domain:    domain,
			Subdomain: result.Value,
			Source:    result.Source,
		}

		var dataJSONBytes []byte

		if dataJSONBytes, err = json.Marshal(data); err != nil {
			return
		}

		if _, err = bw.Write(append(dataJSONBytes, '\n')); err != nil {
			return
		}
	}

	return
}

// Flush pushes the buffered data of writer's stream to the destination.
//
// Parameters:
//   - writer (io.Writer): the destination stream.
//
// Returns:
//   - err (error): non-nil when the flush fails.
func (w *Writer) Flush(writer io.Writer) (err error) {
	bw, ok := w.writers[writer]
	if !ok {
		return
	}

	if err = bw.Flush(); err != nil {
		return
	}

	return
}

func (w *Writer) bufferedWriter(writer io.Writer) (bw *bufio.Writer) {
	bw, ok := w.writers[writer]
	if !ok {
		bw = bufio.NewWriterSize(writer, 256*1024)
		w.writers[writer] = bw
	}

	return
}

type format string

type resultForJSONL struct {
	Domain    string `json:"domain"`
	Subdomain string `json:"subdomain"`
	Source    string `json:"source"`
}

const (
	formatJSONL format = "JSONL"
	formatTXT   format = "TXT"
)

// ErrNoFilePathSpecified is returned when creating a file for an empty
// path.
var ErrNoFilePathSpecified = errors.New("no file path specified")

// NewWriter returns a Writer writing plain text.
func NewWriter() (writer *Writer) {
	writer = &Writer{
		format:  formatTXT,
		writers: make(map[io.Writer]*bufio.Writer),
	}

	return
}
