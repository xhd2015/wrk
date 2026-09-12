package wrkcli

import (
	"bytes"
	"io"
	"sync"
)

// indentingWriter prefixes every line written to w with indent.
// Partial lines keep indent only once at the start of the line.
// Safe for concurrent Write from a single stage's lanes when each lane
// has its own indentingWriter (do not share one across goroutines).
type indentingWriter struct {
	mu     sync.Mutex
	w      io.Writer
	indent string
	atBOL  bool // true = next byte starts a new line (need indent)
}

func newIndentingWriter(w io.Writer, indent string) *indentingWriter {
	if w == nil {
		w = io.Discard
	}
	return &indentingWriter{w: w, indent: indent, atBOL: true}
}

func (iw *indentingWriter) Write(p []byte) (int, error) {
	if iw == nil || len(p) == 0 {
		return len(p), nil
	}
	iw.mu.Lock()
	defer iw.mu.Unlock()

	n := 0
	for len(p) > 0 {
		if iw.atBOL && iw.indent != "" {
			if _, err := io.WriteString(iw.w, iw.indent); err != nil {
				return n, err
			}
			iw.atBOL = false
		}
		i := bytes.IndexByte(p, '\n')
		if i < 0 {
			nw, err := iw.w.Write(p)
			n += nw
			return n, err
		}
		// include newline
		nw, err := iw.w.Write(p[:i+1])
		n += nw
		if err != nil {
			return n, err
		}
		p = p[i+1:]
		iw.atBOL = true
	}
	return n, nil
}

// Flush is a no-op (line-oriented); kept for io.Closer-like call sites.
func (iw *indentingWriter) Flush() error { return nil }
