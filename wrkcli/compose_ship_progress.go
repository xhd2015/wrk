package wrkcli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"regexp"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/xhd2015/dot-pkgs/go-pkgs/terminal/cursor"
	"golang.org/x/term"
)

// Minimal unwind-style progress for compose ship lanes (stderr).
//
// TTY (block): in-place redraw with spinner; clamp lines to term width and
// clear residual rows when the block shrinks.
// Non-TTY (append): one final line per lane on Finish only — no Start/oneline
// appends (avoids duplicate ghost rows like a leftover spinner "-").

type shipProgStatus string

const (
	shipProgWaiting shipProgStatus = "waiting"
	shipProgRunning shipProgStatus = "running"
	shipProgDone    shipProgStatus = "done"
	shipProgFailed  shipProgStatus = "failed"
)

var shipSpinnerFrames = []string{"|", "/", "-", "\\"}

// strip CSI / ESC sequences for visible-width measure.
var shipANSIPattern = regexp.MustCompile(`\x1b\[[0-9;]*[A-Za-z]`)

type shipProgRow struct {
	ID        string
	Label     string
	Status    shipProgStatus
	Oneline   string
	Detail    string
	StartedAt time.Time
	Elapsed   time.Duration
	emitted   bool // append mode: already printed final line
}

type shipProgress struct {
	mu sync.Mutex

	w         io.Writer
	block     bool
	color     bool
	indent    string
	order     []string
	rows      map[string]*shipProgRow
	sinks     map[string]*shipProgSink
	captures  map[string]*bytes.Buffer
	lines     int
	begun     bool
	done      bool
	spinFrame int
	stopSpin  chan struct{}
	spinWG    sync.WaitGroup
	termWidth int
}

func newShipProgress(indent string, color bool, ids []string) *shipProgress {
	w := io.Writer(os.Stderr)
	p := &shipProgress{
		w:         w,
		block:     isShipProgressTTY(w),
		color:     color,
		indent:    indent,
		rows:      make(map[string]*shipProgRow, len(ids)),
		sinks:     make(map[string]*shipProgSink, len(ids)),
		captures:  make(map[string]*bytes.Buffer, len(ids)),
		termWidth: shipTermWidth(w),
		stopSpin:  make(chan struct{}),
	}
	for _, id := range ids {
		p.order = append(p.order, id)
		p.rows[id] = &shipProgRow{ID: id, Label: id, Status: shipProgWaiting}
		buf := &bytes.Buffer{}
		p.captures[id] = buf
		p.sinks[id] = &shipProgSink{p: p, id: id, capture: buf}
	}
	return p
}

func isShipProgressTTY(w io.Writer) bool {
	f, ok := w.(interface{ Fd() uintptr })
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

func shipTermWidth(w io.Writer) int {
	f, ok := w.(interface{ Fd() uintptr })
	if !ok {
		return 80
	}
	width, _, err := term.GetSize(int(f.Fd()))
	if err != nil || width < 40 {
		return 80
	}
	return width
}

func (p *shipProgress) Sink(id string) io.Writer {
	if p == nil {
		return io.Discard
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if s := p.sinks[id]; s != nil {
		return s
	}
	return io.Discard
}

func (p *shipProgress) Capture(id string) string {
	if p == nil {
		return ""
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if buf := p.captures[id]; buf != nil {
		return buf.String()
	}
	return ""
}

func (p *shipProgress) Begin() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.begun || len(p.order) == 0 {
		return
	}
	p.begun = true
	if p.block {
		p.redrawLocked()
		p.spinWG.Add(1)
		go p.spinLoop()
	}
	// Append mode: stay silent until Finish (one line per lane).
}

func (p *shipProgress) spinLoop() {
	defer p.spinWG.Done()
	t := time.NewTicker(120 * time.Millisecond)
	defer t.Stop()
	for {
		select {
		case <-p.stopSpin:
			return
		case <-t.C:
			p.mu.Lock()
			if p.done || !p.block {
				p.mu.Unlock()
				return
			}
			if p.anyRunningLocked() {
				p.spinFrame = (p.spinFrame + 1) % len(shipSpinnerFrames)
				p.redrawLocked()
			}
			p.mu.Unlock()
		}
	}
}

func (p *shipProgress) anyRunningLocked() bool {
	for _, id := range p.order {
		if p.rows[id] != nil && p.rows[id].Status == shipProgRunning {
			return true
		}
	}
	return false
}

func (p *shipProgress) Start(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	row := p.rows[id]
	if row == nil {
		return
	}
	row.Status = shipProgRunning
	row.StartedAt = time.Now()
	row.Elapsed = 0
	if p.block {
		p.emitLocked(id)
	}
	// Append mode: no Start line (avoids ghost spinner rows).
}

func (p *shipProgress) Finish(id string, summary string, err error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	row := p.rows[id]
	if row == nil {
		return
	}
	if !row.StartedAt.IsZero() {
		row.Elapsed = time.Since(row.StartedAt)
	}
	if err != nil {
		row.Status = shipProgFailed
		row.Detail = truncateShipProgress(firstShipLine(err.Error()), 100)
	} else {
		row.Status = shipProgDone
		row.Detail = ""
		if summary != "" {
			row.Oneline = truncateShipProgress(summary, p.onelineBudgetLocked())
		}
	}
	p.emitLocked(id)
}

func (p *shipProgress) Close() {
	p.mu.Lock()
	if p.done {
		p.mu.Unlock()
		return
	}
	p.done = true
	if p.block {
		select {
		case <-p.stopSpin:
		default:
			close(p.stopSpin)
		}
	}
	p.mu.Unlock()
	if p.block {
		p.spinWG.Wait()
		p.mu.Lock()
		if p.lines > 0 {
			p.redrawLocked()
			fmt.Fprintln(p.w)
		}
		p.mu.Unlock()
		return
	}
	// Append mode: emit any lane that never got Finish (cancelled) once.
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, id := range p.order {
		row := p.rows[id]
		if row == nil || row.emitted {
			continue
		}
		if row.Status == shipProgWaiting || row.Status == shipProgRunning {
			row.Status = shipProgFailed
			if row.Detail == "" {
				row.Detail = "cancelled"
			}
		}
		fmt.Fprintln(p.w, p.formatRowLocked(row))
		row.emitted = true
	}
}

func (p *shipProgress) setOneline(id, line string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	row := p.rows[id]
	if row == nil {
		return
	}
	row.Oneline = truncateShipProgress(line, p.onelineBudgetLocked())
	if p.begun && p.block {
		p.emitLocked(id)
	}
	// Append mode: oneline updates are capture-only until Finish.
}

func (p *shipProgress) onelineBudgetLocked() int {
	used := len(p.indent) + 3 + 1 + 2 + 16 + 2 + 8
	budget := p.termWidth - used
	if budget < 16 {
		return 16
	}
	return budget
}

func (p *shipProgress) emitLocked(id string) {
	if !p.begun {
		return
	}
	if p.block {
		p.redrawLocked()
		return
	}
	row := p.rows[id]
	if row == nil || row.emitted {
		return
	}
	// Append mode: only final states (Finish / Close).
	if row.Status != shipProgDone && row.Status != shipProgFailed {
		return
	}
	fmt.Fprintln(p.w, p.formatRowLocked(row))
	row.emitted = true
}

func (p *shipProgress) redrawLocked() {
	old := p.lines
	if old > 0 {
		fmt.Fprint(p.w, cursor.Up(old))
	}
	block := p.renderLocked()
	lines := strings.Split(strings.TrimSuffix(block, "\n"), "\n")
	if block == "" {
		lines = nil
	}
	for _, line := range lines {
		clamped := clampShipProgLine(line, p.termWidth)
		fmt.Fprintf(p.w, "%s%s\n", cursor.ClearLine, cursor.CR+clamped)
	}
	// Clear leftover rows when the block shrinks (wrap / prior taller frame).
	for i := len(lines); i < old; i++ {
		fmt.Fprintf(p.w, "%s%s\n", cursor.ClearLine, cursor.CR)
	}
	if extra := old - len(lines); extra > 0 {
		fmt.Fprint(p.w, cursor.Up(extra))
	}
	p.lines = len(lines)
}

func (p *shipProgress) renderLocked() string {
	var b strings.Builder
	for i, id := range p.order {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(p.formatRowLocked(p.rows[id]))
	}
	return b.String()
}

func (p *shipProgress) formatRowLocked(row *shipProgRow) string {
	if row == nil {
		return p.indent + "   "
	}
	glyph := p.glyphLocked(row)
	rawLabel := row.Label
	pad := 16 - utf8.RuneCountInString(rawLabel)
	if pad < 0 {
		pad = 0
	}
	label := rawLabel + strings.Repeat(" ", pad)
	if row.Status == shipProgDone && p.color {
		label = ansiGrey + ansiStrike + rawLabel + ansiReset + strings.Repeat(" ", pad)
	}
	line := p.indent + "   " + glyph + "  " + label
	tail := row.Oneline
	if row.Status == shipProgFailed && row.Detail != "" {
		tail = row.Detail
	}
	if tail != "" {
		line += "  " + paint(tail, ansiGrey, p.color)
	}
	if elapsed := formatShipProgElapsed(row); elapsed != "" {
		line += "  " + paint(elapsed, ansiGrey, p.color)
	}
	return clampShipProgLine(line, p.termWidth)
}

func formatShipProgElapsed(row *shipProgRow) string {
	if row == nil || row.StartedAt.IsZero() {
		return ""
	}
	var d time.Duration
	switch row.Status {
	case shipProgRunning:
		d = time.Since(row.StartedAt)
	case shipProgDone, shipProgFailed:
		d = row.Elapsed
		if d <= 0 {
			d = time.Since(row.StartedAt)
		}
	default:
		return ""
	}
	return formatShipElapsed(d)
}

func (p *shipProgress) glyphLocked(row *shipProgRow) string {
	switch row.Status {
	case shipProgWaiting:
		return paint("·", ansiGrey, p.color)
	case shipProgRunning:
		return shipSpinnerFrames[p.spinFrame%len(shipSpinnerFrames)]
	case shipProgDone:
		return paint("✓", ansiGreen, p.color)
	case shipProgFailed:
		return paint("✗", ansiRed, p.color)
	default:
		return "?"
	}
}

func truncateShipProgress(s string, max int) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if max < 4 {
		max = 4
	}
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max-3]) + "..."
}

func shipVisibleWidth(s string) int {
	return utf8.RuneCountInString(shipANSIPattern.ReplaceAllString(s, ""))
}

// clampShipProgLine truncates a rendered progress row to term width so TTY
// redraw line counts stay accurate (wrapped rows cause ghost duplicates).
func clampShipProgLine(line string, width int) string {
	if width < 20 {
		width = 20
	}
	if shipVisibleWidth(line) <= width {
		return line
	}
	// Prefer truncating the plain visible tail; keep prefix (indent+glyph+label).
	plain := shipANSIPattern.ReplaceAllString(line, "")
	runes := []rune(plain)
	if len(runes) <= width {
		return line
	}
	return string(runes[:width-3]) + "..."
}

func firstShipLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

type shipProgSink struct {
	p       *shipProgress
	id      string
	mu      sync.Mutex
	partial []byte
	capture *bytes.Buffer
}

func (s *shipProgSink) Write(b []byte) (int, error) {
	if s == nil || len(b) == 0 {
		return len(b), nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.capture != nil {
		_, _ = s.capture.Write(b)
	}
	s.partial = append(s.partial, b...)
	for {
		i := bytes.IndexByte(s.partial, '\n')
		if i < 0 {
			break
		}
		line := string(s.partial[:i])
		s.partial = s.partial[i+1:]
		line = strings.TrimRight(line, "\r")
		if strings.TrimSpace(line) != "" {
			s.p.setOneline(s.id, line)
		}
	}
	if partial := strings.TrimSpace(string(s.partial)); partial != "" {
		s.p.setOneline(s.id, partial)
	}
	return len(b), nil
}
