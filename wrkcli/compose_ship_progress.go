package wrkcli

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"
	"unicode/utf8"

	"github.com/xhd2015/dot-pkgs/go-pkgs/terminal/cursor"
	"golang.org/x/term"
)

// Minimal unwind-style progress for compose ship lanes (stderr).
// Live rows during concurrent apply; full capture available for flush/dump.

type shipProgStatus string

const (
	shipProgWaiting shipProgStatus = "waiting"
	shipProgRunning shipProgStatus = "running"
	shipProgDone    shipProgStatus = "done"
	shipProgFailed  shipProgStatus = "failed"
)

var shipSpinnerFrames = []string{"|", "/", "-", "\\"}

type shipProgRow struct {
	ID        string
	Label     string
	Status    shipProgStatus
	Oneline   string
	Detail    string
	StartedAt time.Time
	Elapsed   time.Duration
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
		return
	}
	fmt.Fprintln(p.w, strings.TrimSuffix(p.renderLocked(), "\n"))
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
	p.emitLocked(id)
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
	if p.begun {
		p.emitLocked(id)
	}
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
	if row == nil {
		return
	}
	fmt.Fprintln(p.w, p.formatRowLocked(row))
}

func (p *shipProgress) redrawLocked() {
	if p.lines > 0 {
		fmt.Fprint(p.w, cursor.Up(p.lines))
	}
	block := p.renderLocked()
	lines := strings.Split(strings.TrimSuffix(block, "\n"), "\n")
	if block == "" {
		lines = nil
	}
	for _, line := range lines {
		fmt.Fprintf(p.w, "%s%s\n", cursor.ClearLine, cursor.CR+line)
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
	return line
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
