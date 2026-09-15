package unwind

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

type actionProgressStatus string

const (
	progWaiting actionProgressStatus = "waiting"
	progRunning actionProgressStatus = "running"
	progDone    actionProgressStatus = "done"
	progFailed  actionProgressStatus = "failed"
	progSkipped actionProgressStatus = "skipped"
)

const (
	ansiStrike = "\x1b[9m"
	spinPeriod = 120 * time.Millisecond
	// Hide/show around an in-place frame so the cursor does not flash
	// mid-rewrite (DEC private modes DECTCEM).
	cursorHide = "\x1b[?25l"
	cursorShow = "\x1b[?25h"
)

var progressSpinnerFrames = []string{"|", "/", "-", "\\"}

// OSC (ESC ] … BEL/ST), CSI, leftover ESC, and C0 (except space).
var (
	progressOSCPattern = regexp.MustCompile(`\x1b\][^\x07\x1b]*(?:\x07|\x1b\\)`)
	progressCSIPattern = regexp.MustCompile(`\x1b\[[0-9;:?]*[A-Za-z]`)
	progressESCPattern = regexp.MustCompile(`\x1b.`)
	progressC0Pattern  = regexp.MustCompile(`[\x00-\x08\x0b\x0c\x0e-\x1f\x7f]`)
)

type actionProgressRow struct {
	ID        string
	Lane      string
	Label     string // action name under the repo group (mode [+ short detail])
	Status    actionProgressStatus
	Oneline   string // rolling last log line
	Detail    string // terminal status detail (fail/skip)
	StartedAt time.Time
	Elapsed   time.Duration // set on Finish; running rows use time.Since(StartedAt)
	Capture   bytes.Buffer
}

type progressGroup struct {
	Lane    string
	Display string
	IDs     []string
}

// actionProgress is an inline (no alt-screen) progress block on stderr.
// On a TTY it redraws in place; otherwise it appends status transitions.
type actionProgress struct {
	mu sync.Mutex

	w      io.Writer
	block  bool
	color  bool
	indent string // kind-column under [n/total]
	order  []string
	rows   map[string]*actionProgressRow
	groups []progressGroup
	sinks  map[string]*actionSink

	lines     int
	begun     bool
	done      bool
	spinFrame int
	stopSpin  chan struct{}
	spinWG    sync.WaitGroup
	termWidth int
}

type progressConfig struct {
	W           io.Writer
	Color       bool
	Indent      string
	LaneDisplay func(lane string) string
}

func newActionProgress(cfg progressConfig, actions []*Action) *actionProgress {
	w := cfg.W
	if w == nil {
		w = io.Discard
	}
	indent := cfg.Indent
	if indent == "" {
		indent = "      " // len("[2/4] ") for total=4
	}
	block := isTerminalWriter(w)
	p := &actionProgress{
		w:         w,
		block:     block,
		color:     cfg.Color,
		indent:    indent,
		rows:      make(map[string]*actionProgressRow, len(actions)),
		sinks:     make(map[string]*actionSink, len(actions)),
		termWidth: terminalWidth(w),
		stopSpin:  make(chan struct{}),
	}

	laneDisp := cfg.LaneDisplay
	if laneDisp == nil {
		laneDisp = func(lane string) string { return lane }
	}

	seenLane := map[string]bool{}
	var laneOrder []string
	for _, a := range actions {
		if a == nil || a.ID == "" {
			continue
		}
		lane := a.Lane
		if lane == "" {
			lane = a.Subject.Display
		}
		if lane == "" {
			lane = "."
		}
		p.order = append(p.order, a.ID)
		p.rows[a.ID] = &actionProgressRow{
			ID:     a.ID,
			Lane:   lane,
			Label:  formatGroupedActionLabel(a),
			Status: progWaiting,
		}
		p.sinks[a.ID] = &actionSink{p: p, id: a.ID}
		if !seenLane[lane] {
			seenLane[lane] = true
			laneOrder = append(laneOrder, lane)
		}
	}
	idsByLane := map[string][]string{}
	for _, id := range p.order {
		idsByLane[p.rows[id].Lane] = append(idsByLane[p.rows[id].Lane], id)
	}
	for _, lane := range laneOrder {
		p.groups = append(p.groups, progressGroup{
			Lane:    lane,
			Display: laneDisp(lane),
			IDs:     idsByLane[lane],
		})
	}
	return p
}

func formatGroupedActionLabel(a *Action) string {
	mode := string(a.Mode)
	if a.Detail != "" && (a.Mode == ModeTagNext || a.Mode == ModeDepUpdate || a.Mode == ModePin) {
		return mode + "  " + truncateProgress(a.Detail, 36)
	}
	return mode
}

func isTerminalWriter(w io.Writer) bool {
	f, ok := w.(interface{ Fd() uintptr })
	if !ok {
		return false
	}
	return term.IsTerminal(int(f.Fd()))
}

func terminalWidth(w io.Writer) int {
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

func truncateProgress(s string, max int) string {
	s = stripProgressControls(s)
	if max < 4 {
		max = 4
	}
	if utf8.RuneCountInString(s) <= max {
		return s
	}
	runes := []rune(s)
	return string(runes[:max-3]) + "..."
}

// ActionSink returns the per-action writer that rolls the oneliner and captures
// full output for failure dumps. Stdout and Stderr should both use this sink.
func (p *actionProgress) ActionSink(id string) io.Writer {
	if p == nil {
		return io.Discard
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if s := p.sinks[id]; s != nil {
		return s
	}
	s := &actionSink{p: p, id: id}
	p.sinks[id] = s
	return s
}

// HostIOFor returns writers for one action (both point at the same sink).
func (p *actionProgress) HostIOFor(id string) HostIO {
	s := p.ActionSink(id)
	return HostIO{Stdout: s, Stderr: s}
}

func (p *actionProgress) Begin(jobs int) {
	_ = jobs
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.begun || len(p.order) == 0 {
		return
	}
	p.begun = true
	if p.block {
		p.redrawLocked() // initial paint; lines tracked for later Up()
		p.spinWG.Add(1)
		go p.spinLoop()
		return
	}
	// Append-only: print the grouped skeleton once (waiting rows).
	fmt.Fprintln(p.w, strings.TrimSuffix(p.renderLocked(), "\n"))
}

func (p *actionProgress) spinLoop() {
	defer p.spinWG.Done()
	t := time.NewTicker(spinPeriod)
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
				p.spinFrame = (p.spinFrame + 1) % len(progressSpinnerFrames)
				p.redrawLocked()
			}
			p.mu.Unlock()
		}
	}
}

func (p *actionProgress) anyRunningLocked() bool {
	for _, id := range p.order {
		if p.rows[id] != nil && p.rows[id].Status == progRunning {
			return true
		}
	}
	return false
}

func (p *actionProgress) Start(id string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	row := p.rows[id]
	if row == nil {
		return
	}
	row.Status = progRunning
	row.StartedAt = time.Now()
	row.Elapsed = 0
	p.emitLocked(id)
}

func (p *actionProgress) Finish(id string, err error) {
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
		if strings.HasPrefix(err.Error(), "skipped:") {
			row.Status = progSkipped
			row.Detail = truncateProgress(err.Error(), 80)
		} else {
			row.Status = progFailed
			row.Detail = truncateProgress(firstLine(err.Error()), 100)
		}
	} else {
		row.Status = progDone
		row.Detail = ""
	}
	p.emitLocked(id)
}

func firstLine(s string) string {
	if i := strings.IndexByte(s, '\n'); i >= 0 {
		return s[:i]
	}
	return s
}

func (p *actionProgress) setOneline(id, line string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	row := p.rows[id]
	if row == nil {
		return
	}
	row.Oneline = truncateProgress(line, p.onelineBudgetLocked())
	if p.begun {
		p.emitLocked(id)
	}
}

func (p *actionProgress) onelineBudgetLocked() int {
	// indent + "   " + glyph + "  " + label(~20) + "  " + elapsed(~8)
	used := len(p.indent) + 3 + 1 + 2 + 20 + 2 + 8
	budget := progressFrameWidth(p.termWidth) - used
	if budget < 16 {
		return 16
	}
	return budget
}

func (p *actionProgress) Close() {
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

// DumpFailures writes captured output for failed actions, kind-aligned.
func (p *actionProgress) DumpFailures() {
	if p == nil {
		return
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	for _, id := range p.order {
		row := p.rows[id]
		if row == nil || row.Status != progFailed {
			continue
		}
		sink := p.sinks[id]
		body := ""
		if sink != nil {
			body = strings.TrimSpace(sink.capture.String())
		}
		if body == "" && row.Detail != "" {
			body = row.Detail
		}
		if body == "" {
			continue
		}
		fmt.Fprintln(p.w, p.indent+"failed "+row.Label+":")
		for _, line := range strings.Split(body, "\n") {
			fmt.Fprintln(p.w, p.indent+line)
		}
	}
}

func (p *actionProgress) emitLocked(id string) {
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
	fmt.Fprintln(p.w, p.formatRowLineLocked(row))
}

func (p *actionProgress) redrawLocked() {
	old := p.lines
	block := p.renderLocked()
	lines := strings.Split(strings.TrimSuffix(block, "\n"), "\n")
	if block == "" {
		lines = nil
	}
	fmt.Fprint(p.w, buildProgressFrame(old, lines, progressFrameWidth(p.termWidth)))
	p.lines = len(lines)
}

func (p *actionProgress) renderLineCountLocked() int {
	n := 0
	for _, g := range p.groups {
		n++ // repo header
		n += len(g.IDs)
	}
	return n
}

func (p *actionProgress) renderLocked() string {
	var b strings.Builder
	first := true
	for _, g := range p.groups {
		if !first {
			b.WriteByte('\n')
		}
		first = false
		b.WriteString(p.indent)
		b.WriteString(g.Display)
		for _, id := range g.IDs {
			b.WriteByte('\n')
			b.WriteString(p.formatRowLineLocked(p.rows[id]))
		}
	}
	return b.String()
}

func (p *actionProgress) formatRowLineLocked(row *actionProgressRow) string {
	if row == nil {
		return p.indent + "   "
	}
	glyph := p.glyphLocked(row)
	label := row.Label
	if row.Status == progDone && p.color {
		label = ansiGrey + ansiStrike + row.Label + ansiReset
	}
	line := p.indent + "   " + glyph + "  " + label
	tail := row.Oneline
	if row.Status == progFailed || row.Status == progSkipped {
		if row.Detail != "" {
			tail = row.Detail
		}
	}
	if tail != "" {
		if p.color {
			tail = paint(tail, ansiGrey, true)
		}
		line += "  " + tail
	}
	if elapsed := formatActionElapsed(row); elapsed != "" {
		line += "  " + paint(elapsed, ansiGrey, p.color)
	}
	return clampProgressLine(line, progressFrameWidth(p.termWidth))
}

// formatActionElapsed returns a compact duration for running/finished rows.
// Waiting rows stay blank. Shape: 850ms / 4.8s / 1m02s.
func formatActionElapsed(row *actionProgressRow) string {
	if row == nil || row.StartedAt.IsZero() {
		return ""
	}
	var d time.Duration
	switch row.Status {
	case progRunning:
		d = time.Since(row.StartedAt)
	case progDone, progFailed, progSkipped:
		d = row.Elapsed
		if d <= 0 {
			d = time.Since(row.StartedAt)
		}
	default:
		return ""
	}
	return formatProgressElapsed(d)
}

func formatProgressElapsed(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	if d < time.Second {
		ms := d.Milliseconds()
		if ms < 1 && d > 0 {
			ms = 1
		}
		return fmt.Sprintf("%dms", ms)
	}
	if d < time.Minute {
		// One decimal place for sub-minute seconds.
		sec := float64(d) / float64(time.Second)
		return fmt.Sprintf("%.1fs", sec)
	}
	mins := int(d / time.Minute)
	secs := int((d % time.Minute) / time.Second)
	return fmt.Sprintf("%dm%02ds", mins, secs)
}

func (p *actionProgress) glyphLocked(row *actionProgressRow) string {
	switch row.Status {
	case progWaiting:
		g := "·"
		return paint(g, ansiGrey, p.color)
	case progRunning:
		g := progressSpinnerFrames[p.spinFrame%len(progressSpinnerFrames)]
		return g
	case progDone:
		g := "✓"
		return paint(g, ansiGreen, p.color)
	case progFailed:
		g := "✗"
		return paint(g, ansiRed, p.color)
	case progSkipped:
		g := "-"
		return paint(g, ansiGrey, p.color)
	default:
		return "?"
	}
}

// actionSink collects bytes for one action: rolls oneliner + full capture.
type actionSink struct {
	p       *actionProgress
	id      string
	mu      sync.Mutex
	partial []byte
	capture bytes.Buffer
}

func (s *actionSink) Write(b []byte) (int, error) {
	if s == nil || len(b) == 0 {
		return len(b), nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	_, _ = s.capture.Write(b)
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

// Ensure actionSink implements io.Writer.
var _ io.Writer = (*actionSink)(nil)

// progressFrameWidth is the max visible columns for one progress row.
// Subtract 1 for xenl: a full-width line plus LF can occupy two visual rows.
func progressFrameWidth(termWidth int) int {
	w := termWidth - 1
	if w < 1 {
		return 1
	}
	return w
}

// stripProgressControls removes cursor-moving / query sequences from untrusted
// action output so reprinting a row cannot desync CSI nA.
func stripProgressControls(s string) string {
	s = strings.ReplaceAll(s, "\r\n", "\n")
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	s = progressOSCPattern.ReplaceAllString(s, "")
	s = progressCSIPattern.ReplaceAllString(s, "")
	s = progressESCPattern.ReplaceAllString(s, "")
	s = progressC0Pattern.ReplaceAllString(s, "")
	return strings.TrimSpace(s)
}

func progressVisibleWidth(s string) int {
	s = progressOSCPattern.ReplaceAllString(s, "")
	s = progressCSIPattern.ReplaceAllString(s, "")
	return utf8.RuneCountInString(s)
}

// clampProgressLine truncates a rendered progress row to width so TTY
// redraw line counts stay accurate (wrapped rows cause ghost duplicates).
func clampProgressLine(line string, width int) string {
	if width < 1 {
		width = 1
	}
	if progressVisibleWidth(line) <= width {
		return line
	}
	plain := progressOSCPattern.ReplaceAllString(line, "")
	plain = progressCSIPattern.ReplaceAllString(plain, "")
	runes := []rune(plain)
	if len(runes) <= width {
		return line
	}
	if width <= 3 {
		return string(runes[:width])
	}
	return string(runes[:width-3]) + "..."
}

// buildProgressFrame is one in-place TTY rewrite: hide cursor, Up(old),
// CR+EL+clamped line per row, clear leftover rows, show cursor.
func buildProgressFrame(old int, lines []string, width int) string {
	var b strings.Builder
	b.WriteString(cursorHide)
	if old > 0 {
		b.WriteString(cursor.Up(old))
	}
	for _, line := range lines {
		b.WriteString(cursor.ClearLine)
		b.WriteString(cursor.CR)
		b.WriteString(clampProgressLine(line, width))
		b.WriteByte('\n')
	}
	extra := old - len(lines)
	if extra < 0 {
		extra = 0
	}
	for i := 0; i < extra; i++ {
		b.WriteString(cursor.ClearLine)
		b.WriteString(cursor.CR)
		b.WriteByte('\n')
	}
	if extra > 0 {
		b.WriteString(cursor.Up(extra))
	}
	b.WriteString(cursorShow)
	return b.String()
}

// resolveProgressColor mirrors stderr color policy for the progress block.
func resolveProgressColor(forceColor, noColor bool) bool {
	if noColor {
		return false
	}
	if forceColor {
		return true
	}
	if os.Getenv("NO_COLOR") != "" {
		return false
	}
	return term.IsTerminal(int(os.Stderr.Fd()))
}
