package wrkcli

import (
	"fmt"
	"io"
	"os"
	"strings"
)

// composeStageWriter prints unwind-style [n/N] stage markers on stderr (flush
// left) and exposes kind-aligned Out/Err writers for all body output under the
// open stage (go-best-practice cli/output/staged-markers).
type composeStageWriter struct {
	total    int
	cur      int
	indent   string
	colorErr bool
	colorOut bool
	errW     io.Writer
	outW     io.Writer

	bodyOut *indentingWriter
	bodyErr *indentingWriter
}

func newComposeStageWriter(total int, colorErr, colorOut bool) *composeStageWriter {
	if total < 1 {
		total = 1
	}
	indent := strings.Repeat(" ", len(fmt.Sprintf("[%d/%d] ", total, total)))
	s := &composeStageWriter{
		total:    total,
		indent:   indent,
		colorErr: colorErr,
		colorOut: colorOut,
		errW:     os.Stderr,
		outW:     os.Stdout,
	}
	s.bodyOut = newIndentingWriter(s.outW, indent)
	s.bodyErr = newIndentingWriter(s.errW, indent)
	return s
}

func (s *composeStageWriter) mark(i int, name string) {
	if s == nil {
		return
	}
	s.cur = i
	// Markers stay flush left.
	fmt.Fprintf(s.errW, "[%d/%d] %s\n", i, s.total, name)
}

func (s *composeStageWriter) detail(msg string) {
	if s == nil {
		return
	}
	fmt.Fprintln(s.Err(), paint(msg, ansiGrey, s.colorErr))
}

func (s *composeStageWriter) line(msg string) {
	if s == nil {
		fmt.Fprintln(os.Stdout, msg)
		return
	}
	fmt.Fprintln(s.Out(), msg)
}

// Out returns a kind-aligned stdout writer for the open stage body.
func (s *composeStageWriter) Out() io.Writer {
	if s == nil || s.bodyOut == nil {
		return os.Stdout
	}
	return s.bodyOut
}

// Err returns a kind-aligned stderr writer for the open stage body.
func (s *composeStageWriter) Err() io.Writer {
	if s == nil || s.bodyErr == nil {
		return os.Stderr
	}
	return s.bodyErr
}

// Indent is the kind-column pad (spaces matching "[n/N] ").
func (s *composeStageWriter) Indent() string {
	if s == nil {
		return ""
	}
	return s.indent
}

// newLandComposeStages builds a stage writer for done/merge-back compose.
// Stages (only enabled ones count): commit → land → ship → exec.
// Returns the writer and the 1-based index of the ship stage (0 if no ship).
func newLandComposeStages(hasCommit, hasShip, hasExec bool, colorFlag, noColorFlag bool) (*composeStageWriter, int) {
	total := 0
	if hasCommit {
		total++
	}
	total++ // land always
	shipIdx := 0
	if hasShip {
		total++
		shipIdx = total
	}
	if hasExec {
		total++
	}
	st := newComposeStageWriter(total, resolveStderrColor(colorFlag, noColorFlag), resolveStdoutColor(colorFlag, noColorFlag))
	return st, shipIdx
}
