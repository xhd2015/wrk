package wrkcli

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Focusable rows for ↑/↓ (one stop per row — not per Run chip).
type dashFocusKind int

const (
	focusStage  dashFocusKind = iota // pre/after stage row (Enter/space = toggle)
	focusMain                        // MERGE BACK / DONE radio
	focusRunAll                      // batch RUN ALL
	focusCancel
)

// dashFocus is one keyboard focus row.
type dashFocus struct {
	kind    dashFocusKind
	stageID string
	label   string
}

// dashHit maps a view cell range for mouse.
// If runStage is non-empty, click runs that stage alone (does not change row focus list).
// Otherwise click focuses `focus` and activates Enter semantics.
type dashHit struct {
	y0, y1    int
	x0, x1    int // if x1<=x0, whole line matches
	focus     int
	runStage  string // single-phase run id; exclusive with normal activate when set
}

// teaDashModel is the Bubble Tea model for bare `wrk` on a real TTY.
type teaDashModel struct {
	workDir string
	status  string

	addAll         bool
	addDisabled    bool
	genCommitMsg   bool
	commit         bool
	agentRunner    string
	primaryDone    bool // true → DONE; false → MERGE BACK (default, safer)
	mainDisabled   bool
	sync           bool
	tagNext        bool
	push           bool
	reinstallLocal bool

	focus  []dashFocus
	cursor int
	width  int
	height int // terminal rows (WindowSizeMsg); for mouse Y mapping without alt-screen
	hitmap []dashHit
	viewLines int // last rendered line count (hitmap local Y is 0..viewLines-1)
	color  bool   // ANSI styles when true

	// loadingID: non-empty while an op runs in-process (no UI tear-down / flash).
	// Values: stage ids ("add-changes", "sync", …) or "run-all".
	loadingID    string
	loadingFrame int // spinner frame index

	quitOutcome int // only cancel leaves the program
}

const (
	teaOutcomeNone   = 0
	teaOutcomeCancel = 1
)

// dashOpDoneMsg is delivered when an in-TUI background op finishes.
type dashOpDoneMsg struct {
	loadingID string
	status    string
	err       error
}

// dashTickMsg advances the loading spinner while an op runs.
type dashTickMsg struct{}

// ASCII spinner keeps column width stable under dashPad (single-byte cells).
var dashSpinnerFrames = []string{"|", "/", "-", "\\"}

// Dashboard color roles (TTY only; honor NO_COLOR).
const (
	ansiCyan      = "\x1b[36m"
	ansiBoldGreen = "\x1b[1;32m"
	ansiBoldCyan  = "\x1b[1;36m"
	ansiBoldYellow = "\x1b[1;33m"
	ansiYellow    = "\x1b[33m"
)

func dashColorOn() bool {
	return os.Getenv("NO_COLOR") == ""
}

// dashMouseIsLeftClick reports a primary-button click we should handle once.
// Handle Press only (not Release) to avoid double-toggle; release often has Button=None.
// Accept Button=Left or legacy Type=MouseLeft.
func dashMouseIsLeftClick(msg tea.MouseMsg) bool {
	ev := tea.MouseEvent(msg)
	if ev.IsWheel() {
		return false
	}
	if msg.Action != tea.MouseActionPress {
		return false
	}
	return msg.Button == tea.MouseButtonLeft || msg.Type == tea.MouseLeft
}

// mapMouseY converts terminal-absolute mouse Y to view-local line index.
// Without alt-screen, Bubble Tea paints inline and typically sits at the bottom
// of the visible terminal: originY ≈ height - viewLines.
// Also tries raw Y as fallback when the mapped coordinate misses (UI not bottom-anchored).
func (m *teaDashModel) mapMouseY(absY int) (local int, ok bool) {
	if m.height > 0 && m.viewLines > 0 {
		origin := m.height - m.viewLines
		if origin < 0 {
			origin = 0
		}
		local = absY - origin
		if local >= 0 && local < m.viewLines {
			return local, true
		}
	}
	// Fallback: treat absolute Y as local (works if UI starts at row 0).
	if absY >= 0 && (m.viewLines == 0 || absY < m.viewLines) {
		return absY, true
	}
	return absY, false
}

func (m *teaDashModel) hitTest(x, y int) (h dashHit, ok bool) {
	for _, cand := range m.hitmap {
		if y < cand.y0 || y >= cand.y1 {
			continue
		}
		if cand.x1 > cand.x0 && (x < cand.x0 || x >= cand.x1) {
			continue
		}
		return cand, true
	}
	return dashHit{}, false
}

func dashPaint(on bool, code, s string) string {
	if !on || s == "" || code == "" {
		return s
	}
	return code + s + ansiReset
}

func newTeaDashModel(workDir, status string) teaDashModel {
	addable := dashboardHasAddableDirt(workDir)
	onMain := dashboardIsMainCheckout(workDir)
	m := teaDashModel{
		workDir:        workDir,
		status:         status,
		addAll:         addable,
		addDisabled:    !addable,
		genCommitMsg:   true,
		commit:         true,
		agentRunner:    "commandcode",
		primaryDone:    false, // MERGE BACK default
		mainDisabled:   onMain,
		sync:           true,
		tagNext:        true,
		push:           true,
		reinstallLocal: true,
		color:          dashColorOn(),
	}
	m.rebuildFocus()
	m.cursor = m.firstEnabledFocus()
	return m
}

func (m *teaDashModel) rebuildFocus() {
	// Row-wise only: arrows never land on per-row [ Run ] chips.
	m.focus = nil
	for _, id := range []string{"add-changes", "gen-commit-msg", "commit"} {
		m.focus = append(m.focus, dashFocus{kind: focusStage, stageID: id, label: id})
	}
	m.focus = append(m.focus,
		dashFocus{kind: focusMain, stageID: "merge-back", label: "MERGE BACK"},
		dashFocus{kind: focusMain, stageID: "done", label: "DONE"},
	)
	for _, id := range []string{"sync", "tag-next", "push", "reinstall-local"} {
		m.focus = append(m.focus, dashFocus{kind: focusStage, stageID: id, label: id})
	}
	m.focus = append(m.focus,
		dashFocus{kind: focusRunAll, stageID: "run-all", label: "RUN ALL"},
		dashFocus{kind: focusCancel, stageID: "cancel", label: "CANCEL"},
	)
	if m.cursor >= len(m.focus) {
		m.cursor = len(m.focus) - 1
	}
	if m.cursor < 0 {
		m.cursor = 0
	}
}

func (m *teaDashModel) firstEnabledFocus() int {
	// Prefer Main (MERGE BACK) so accidental Enter does not run "add changes".
	for i, f := range m.focus {
		if f.kind == focusMain && !m.focusDisabled(f) {
			return i
		}
	}
	for i, f := range m.focus {
		if !m.focusDisabled(f) {
			return i
		}
	}
	return 0
}

// clearInlineDashboard erases the previous inline (non-alt-screen) Bubble Tea
// frame so the next program does not stack a second copy underneath.
// Cursor is assumed just below the last line of the previous view (tea default).
func clearInlineDashboard(viewLines int) {
	if viewLines <= 0 {
		return
	}
	// Move to the first line of the previous frame, then erase to end of screen.
	// CSI n A = cursor up n; CSI 0 J = erase from cursor to end of display.
	fmt.Fprintf(os.Stdout, "\x1b[%dA\x1b[0J", viewLines)
}

func (m *teaDashModel) focusDisabled(f dashFocus) bool {
	switch f.kind {
	case focusStage:
		// Stage rows stay focusable even when add-changes is disabled (toggle no-ops).
		return false
	case focusMain:
		return m.mainDisabled
	default:
		return false
	}
}

func (m *teaDashModel) stageRunDisabled(id string) bool {
	return id == "add-changes" && m.addDisabled
}

func (m *teaDashModel) Init() tea.Cmd { return nil }

func (m *teaDashModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		return m, nil
	case dashTickMsg:
		if m.loadingID == "" {
			return m, nil
		}
		m.loadingFrame = (m.loadingFrame + 1) % len(dashSpinnerFrames)
		return m, dashSpinnerTick()
	case dashOpDoneMsg:
		m.loadingID = ""
		m.loadingFrame = 0
		if msg.err != nil {
			m.status = "error: " + msg.err.Error()
		} else if msg.status != "" {
			m.status = msg.status
		} else {
			m.status = "ok"
		}
		if len(m.status) > 90 {
			m.status = m.status[:87] + "..."
		}
		// Refresh dirt gates after staging/compose.
		m.addDisabled = !dashboardHasAddableDirt(m.workDir)
		if m.addDisabled {
			m.addAll = false
		}
		m.mainDisabled = dashboardIsMainCheckout(m.workDir)
		return m, nil
	case tea.MouseMsg:
		if m.loadingID != "" {
			return m, nil // ignore clicks while running
		}
		if !dashMouseIsLeftClick(msg) {
			return m, nil
		}
		// No alt-screen: map absolute mouse Y → view-local line (see mapMouseY).
		x := msg.X
		yLocal, _ := m.mapMouseY(msg.Y)
		h, ok := m.hitTest(x, yLocal)
		if !ok {
			// Retry absolute Y (UI starting at top of viewport).
			h, ok = m.hitTest(x, msg.Y)
		}
		if !ok {
			return m, nil
		}
		if h.runStage != "" {
			return m.startStageRun(h.runStage)
		}
		if h.focus >= 0 {
			m.cursor = h.focus
			return m.activateFocus()
		}
		return m, nil
	case tea.KeyMsg:
		// Allow quit even while loading; block other input during ops.
		switch msg.String() {
		case "ctrl+c", "q", "esc":
			if m.loadingID != "" {
				// Soft cancel request: still leave (in-flight op may finish in background).
				m.quitOutcome = teaOutcomeCancel
				return m, tea.Quit
			}
			m.quitOutcome = teaOutcomeCancel
			return m, tea.Quit
		}
		if m.loadingID != "" {
			return m, nil
		}
		switch msg.String() {
		case "up", "k":
			m.moveFocus(-1)
			return m, nil
		case "down", "j":
			m.moveFocus(1)
			return m, nil
		case " ":
			m.applySpace()
			return m, nil
		case "enter":
			return m.activateFocus()
		}
	}
	return m, nil
}

func dashSpinnerTick() tea.Cmd {
	return tea.Tick(120*time.Millisecond, func(time.Time) tea.Msg {
		return dashTickMsg{}
	})
}

// dashRunOpCmd runs a dashboard op off the UI thread; stdout/stderr captured
// so compose logs do not corrupt the inline TUI.
func dashRunOpCmd(workDir, loadingID string, addOnly bool, recipe dashboardRecipe) tea.Cmd {
	return func() tea.Msg {
		status, err := runDashboardOpCaptured(workDir, addOnly, recipe)
		return dashOpDoneMsg{loadingID: loadingID, status: status, err: err}
	}
}

// runDashboardOpCaptured executes git-add or compose with stdio redirected to a pipe.
func runDashboardOpCaptured(workDir string, addOnly bool, recipe dashboardRecipe) (status string, err error) {
	oldOut, oldErr := os.Stdout, os.Stderr
	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		return "", pipeErr
	}
	os.Stdout, os.Stderr = w, w
	done := make(chan struct{})
	var buf strings.Builder
	go func() {
		_, _ = io.Copy(&buf, r)
		close(done)
	}()

	runErr := error(nil)
	if addOnly {
		runErr = gitRunDir(workDir, "add", "-A")
		if runErr == nil {
			status = "ok  staged: git add -A"
		}
	} else {
		runErr = runDashboardComposeWithRecipe(workDir, nil, recipe)
		if runErr == nil {
			argv := strings.Join(composeArgvFromRecipe(recipe), " ")
			status = "ok  " + argv
		}
	}

	_ = w.Close()
	<-done
	os.Stdout, os.Stderr = oldOut, oldErr
	_ = r.Close()
	_ = buf.String() // discarded; status line carries summary
	if runErr != nil {
		return "", runErr
	}
	if len(status) > 90 {
		status = status[:87] + "..."
	}
	return status, nil
}

func (m *teaDashModel) moveFocus(delta int) {
	if len(m.focus) == 0 {
		return
	}
	n := len(m.focus)
	start := m.cursor
	for i := 0; i < n; i++ {
		m.cursor = (m.cursor + delta + n) % n
		if !m.focusDisabled(m.focus[m.cursor]) {
			return
		}
	}
	m.cursor = start
}

func (m *teaDashModel) applySpace() {
	if m.loadingID != "" {
		return
	}
	if m.cursor < 0 || m.cursor >= len(m.focus) {
		return
	}
	f := m.focus[m.cursor]
	switch f.kind {
	case focusStage:
		if f.stageID == "add-changes" && m.addDisabled {
			return
		}
		m.toggleStage(f.stageID)
	case focusMain:
		if !m.mainDisabled {
			m.selectMain(f.stageID)
		}
	}
}

// startStageRun runs one stage in-process with a loading spinner (no UI tear-down).
func (m *teaDashModel) startStageRun(stageID string) (tea.Model, tea.Cmd) {
	if m.loadingID != "" || m.stageRunDisabled(stageID) {
		return m, nil
	}
	m.loadingID = stageID
	m.loadingFrame = 0
	m.status = "running…"
	if stageID == "add-changes" {
		return m, tea.Batch(
			dashRunOpCmd(m.workDir, stageID, true, dashboardRecipe{}),
			dashSpinnerTick(),
		)
	}
	r, ok := m.singleStageRecipe(stageID)
	if !ok {
		m.loadingID = ""
		return m, nil
	}
	return m, tea.Batch(
		dashRunOpCmd(m.workDir, stageID, false, r),
		dashSpinnerTick(),
	)
}

func (m *teaDashModel) startRunAll() (tea.Model, tea.Cmd) {
	if m.loadingID != "" {
		return m, nil
	}
	m.loadingID = "run-all"
	m.loadingFrame = 0
	m.status = "running…"
	r := m.toRecipe()
	return m, tea.Batch(
		dashRunOpCmd(m.workDir, "run-all", false, r),
		dashSpinnerTick(),
	)
}

func (m *teaDashModel) activateFocus() (tea.Model, tea.Cmd) {
	if m.loadingID != "" {
		return m, nil
	}
	if m.cursor < 0 || m.cursor >= len(m.focus) {
		return m, nil
	}
	f := m.focus[m.cursor]
	switch f.kind {
	case focusStage:
		// Enter on stage row = single-phase run (space toggles batch include).
		return m.startStageRun(f.stageID)
	case focusMain:
		if !m.mainDisabled {
			m.selectMain(f.stageID)
		}
		return m, nil
	case focusRunAll:
		return m.startRunAll()
	case focusCancel:
		m.quitOutcome = teaOutcomeCancel
		return m, tea.Quit
	}
	return m, nil
}

func (m *teaDashModel) spinnerGlyph() string {
	if len(dashSpinnerFrames) == 0 {
		return "…"
	}
	return dashSpinnerFrames[m.loadingFrame%len(dashSpinnerFrames)]
}

func (m *teaDashModel) toggleStage(id string) {
	switch id {
	case "add-changes":
		if !m.addDisabled {
			m.addAll = !m.addAll
		}
	case "gen-commit-msg":
		m.genCommitMsg = !m.genCommitMsg
	case "commit":
		m.commit = !m.commit
	case "sync":
		m.sync = !m.sync
	case "tag-next":
		m.tagNext = !m.tagNext
	case "push":
		m.push = !m.push
	case "reinstall-local":
		m.reinstallLocal = !m.reinstallLocal
	}
}

func (m *teaDashModel) selectMain(id string) {
	if m.mainDisabled {
		return
	}
	switch id {
	case "done":
		m.primaryDone = true
	case "merge-back":
		m.primaryDone = false
	}
}

func (m *teaDashModel) glyph(id string) string {
	on := func(v bool) string {
		if v {
			return dashGlyphOn
		}
		return dashGlyphOff
	}
	switch id {
	case "add-changes":
		if m.addDisabled {
			return dashGlyphDisabled
		}
		return on(m.addAll)
	case "gen-commit-msg":
		return on(m.genCommitMsg)
	case "commit":
		return on(m.commit)
	case "done":
		if m.mainDisabled {
			return dashGlyphDisabled
		}
		return on(m.primaryDone)
	case "merge-back":
		if m.mainDisabled {
			return dashGlyphDisabled
		}
		return on(!m.primaryDone)
	case "sync":
		return on(m.sync)
	case "tag-next":
		return on(m.tagNext)
	case "push":
		return on(m.push)
	case "reinstall-local":
		return on(m.reinstallLocal)
	default:
		return dashGlyphOff
	}
}

func (m *teaDashModel) singleStageRecipe(id string) (dashboardRecipe, bool) {
	r := dashboardRecipe{agentRunner: m.agentRunner, dryRun: os.Getenv(envDashboardDryRun) == "1"}
	switch id {
	case "gen-commit-msg":
		r.genCommitMsg = true
		r.addAll = m.addAll && !m.addDisabled
		return r, true
	case "commit":
		r.genCommitMsg = true
		r.addAll = m.addAll && !m.addDisabled
		r.commit = true
		return r, true
	case "sync":
		r.sync = true
		return r, true
	case "tag-next":
		r.tagNext = true
		return r, true
	case "push":
		r.push = true
		return r, true
	case "reinstall-local":
		r.reinstallLocal = true
		return r, true
	default:
		return r, false
	}
}

func (m *teaDashModel) toRecipe() dashboardRecipe {
	return dashboardRecipe{
		addAll:         m.addAll && !m.addDisabled,
		genCommitMsg:   m.genCommitMsg,
		commit:         m.commit,
		agentRunner:    m.agentRunner,
		done:           m.primaryDone && !m.mainDisabled,
		mergeBack:      !m.primaryDone && !m.mainDisabled,
		sync:           m.sync,
		tagNext:        m.tagNext,
		push:           m.push,
		reinstallLocal: m.reinstallLocal,
		dryRun:         os.Getenv(envDashboardDryRun) == "1",
	}
}

func (m *teaDashModel) isFocused(kind dashFocusKind, stageID string) bool {
	if m.cursor < 0 || m.cursor >= len(m.focus) {
		return false
	}
	f := m.focus[m.cursor]
	return f.kind == kind && f.stageID == stageID
}

func (m *teaDashModel) isStageRowFocused(id string) bool {
	return m.isFocused(focusStage, id)
}

func (m *teaDashModel) View() string {
	return m.renderView()
}

// --- compact layout helpers ---

func dashLayoutWidth(termW int) int {
	w := termW
	if w < 56 {
		w = 56
	}
	if w > 72 {
		w = 72
	}
	return w
}

func dashPad(s string, n int) string {
	// Pad/truncate by display-ish length (ASCII UI; glyphs are 3 runes).
	if n <= 0 {
		return ""
	}
	r := []rune(s)
	if len(r) >= n {
		return string(r[:n])
	}
	return s + strings.Repeat(" ", n-len(r))
}

func dashRuneLen(s string) int {
	return len([]rune(s))
}

// frameLine builds a full-width line: │ + pad(inner) + │  (width W).
func frameLine(w int, inner string) string {
	innerW := w - 2
	if innerW < 1 {
		innerW = 1
	}
	return "│" + dashPad(inner, innerW) + "│"
}

func frameTop(w int, title string) string {
	// ┌─ title ──…─┐
	// title includes leading space after ─
	body := "─ " + title + " "
	rest := w - 2 - dashRuneLen(body)
	if rest < 0 {
		body = dashPad(body, w-2)
		rest = 0
	}
	return "┌" + body + strings.Repeat("─", rest) + "┐"
}

func frameSection(w int, name string) string {
	body := "─ " + name + " "
	rest := w - 2 - dashRuneLen(body)
	if rest < 0 {
		body = dashPad(body, w-2)
		rest = 0
	}
	return "├" + body + strings.Repeat("─", rest) + "┤"
}

func frameBottom(w int) string {
	return "└" + strings.Repeat("─", w-2) + "┘"
}

func (m *teaDashModel) paintGlyph(g string) string {
	switch g {
	case dashGlyphOn:
		return dashPaint(m.color, ansiGreen, g)
	case dashGlyphDisabled:
		return dashPaint(m.color, ansiGrey, g)
	default:
		return dashPaint(m.color, ansiGrey, g)
	}
}

func (m *teaDashModel) paintStatus(st string) string {
	plain := " status  " + st
	if !m.color {
		return plain
	}
	if strings.HasPrefix(st, "error:") || strings.HasPrefix(st, "error ") {
		return " status  " + dashPaint(true, ansiRed, st)
	}
	if strings.HasPrefix(st, "ok") {
		return " status  " + dashPaint(true, ansiGreen, st)
	}
	return " status  " + dashPaint(true, ansiGrey, st)
}

func (m *teaDashModel) renderView() string {
	var hits []dashHit
	var lines []string

	w := dashLayoutWidth(m.width)
	innerW := w - 2
	// Run chip column: last 10 cells of content (inside borders).
	runW := 10
	if innerW < 40 {
		runW = 8
	}
	leftW := innerW - runW
	if leftW < 12 {
		leftW = innerW / 2
		runW = innerW - leftW
	}

	// Content columns inside borders: x=1 .. w-2 (border at 0 and w-1).
	// Toggle hit: x 1 .. 1+leftW
	// Run hit: x 1+leftW .. w-1
	addPlain := func(s string) {
		lines = append(lines, s)
	}

	focusIdx := func(kind dashFocusKind, id string) int {
		for i, f := range m.focus {
			if f.kind == kind && f.stageID == id {
				return i
			}
		}
		return -1
	}

	// Border lines (gray)
	paintBorder := func(s string) string {
		return dashPaint(m.color, ansiGrey, s)
	}

	y := 0
	titleInner := "wrk dashboard"
	top := frameTop(w, titleInner)
	addPlain(paintBorder(top))
	y++

	st := strings.TrimSpace(m.status)
	if st == "" {
		st = "ready"
	}
	statusLine := paintBorder("│") + dashPad(m.paintStatus(st), innerW) + paintBorder("│")
	addPlain(statusLine)
	y++

	addSection := func(name string) {
		addPlain(paintBorder(frameSection(w, name)))
		y++
	}

	// Stage row: one line; ↑/↓ focuses the row; [ Run ] is click/`r` only.
	renderStage := func(id, title, extra string) {
		ti := focusIdx(focusStage, id)
		rowF := m.isStageRowFocused(id)
		g := m.glyph(id)

		left := g + " " + title
		if extra != "" {
			left += "  " + extra
		}
		if rowF {
			left = "▶" + left
		} else {
			left = " " + left
		}

		runLabel := "[ Run ]"
		loadingHere := m.loadingID == id
		if m.stageRunDisabled(id) {
			runLabel = "[  —  ]"
		} else if loadingHere {
			runLabel = "[ " + m.spinnerGlyph() + " ]"
		}

		runPlain := strings.Repeat(" ", max(0, runW-dashRuneLen(runLabel))) + runLabel
		if dashRuneLen(runPlain) > runW {
			runPlain = dashPad(runLabel, runW)
		}

		gCol := m.paintGlyph(g)
		titlePart := title
		extraPart := extra
		if m.color {
			if g == dashGlyphDisabled || (id == "add-changes" && m.addDisabled) {
				titlePart = dashPaint(true, ansiGrey, title)
				extraPart = dashPaint(true, ansiGrey, extra)
			} else {
				extraPart = dashPaint(true, ansiGrey, extra)
			}
			if rowF && !loadingHere {
				titlePart = dashPaint(true, ansiBoldGreen, title)
			}
		}
		leftVis := " "
		if rowF {
			leftVis = dashPaint(m.color, ansiGreen, "▶")
		}
		leftVis += gCol + " " + titlePart
		if extra != "" {
			leftVis += "  " + extraPart
		}
		padN := leftW - dashRuneLen(left)
		if padN < 0 {
			padN = 0
		}
		leftVis = leftVis + strings.Repeat(" ", padN)

		runVis := runPlain
		if m.stageRunDisabled(id) {
			runVis = dashPaint(m.color, ansiGrey, runPlain)
		} else if loadingHere {
			runVis = dashPaint(m.color, ansiBoldYellow, runPlain)
		} else {
			runVis = dashPaint(m.color, ansiCyan, runPlain)
		}

		line := paintBorder("│") + leftVis + runVis + paintBorder("│")
		// Left: select/toggle row; right: single-phase Run (mouse only for chip).
		if ti >= 0 {
			hits = append(hits, dashHit{y0: y, y1: y + 1, x0: 1, x1: 1 + leftW, focus: ti})
		}
		if !m.stageRunDisabled(id) {
			hits = append(hits, dashHit{y0: y, y1: y + 1, x0: 1 + leftW, x1: w - 1, focus: -1, runStage: id})
		}
		lines = append(lines, line)
		y++
	}

	addSection("Pre")
	renderStage("add-changes", "add changes", "--add-all")
	renderStage("gen-commit-msg", "gen-commit-msg", "agent="+m.agentRunner)
	renderStage("commit", "commit", "--commit")

	addSection("Main")
	for _, id := range []string{"merge-back", "done"} {
		fi := focusIdx(focusMain, id)
		foc := m.isFocused(focusMain, id)
		sel := (id == "merge-back" && !m.primaryDone) || (id == "done" && m.primaryDone)
		name := "MERGE BACK"
		desc := "merge, keep worktree"
		if id == "done" {
			name = "DONE"
			desc = "merge + remove worktree"
		}
		g := m.glyph(id)
		prefix := " "
		if foc {
			prefix = "▶"
		}
		body := prefix + g + " " + name + "   " + desc
		bodyPlain := dashPad(body, innerW)

		gCol := m.paintGlyph(g)
		nameCol := name
		descCol := dashPaint(m.color, ansiGrey, desc)
		if m.mainDisabled {
			nameCol = dashPaint(m.color, ansiGrey, name)
			descCol = dashPaint(m.color, ansiGrey, desc)
		} else if sel {
			nameCol = dashPaint(m.color, ansiBoldGreen, name)
		}
		if foc {
			prefix = dashPaint(m.color, ansiGreen, "▶")
		} else {
			prefix = " "
		}
		vis := prefix + gCol + " " + nameCol + "   " + descCol
		padN := innerW - dashRuneLen(body)
		if padN < 0 {
			padN = 0
		}
		vis += strings.Repeat(" ", padN)
		_ = bodyPlain
		line := paintBorder("│") + vis + paintBorder("│")
		if fi >= 0 {
			hits = append(hits, dashHit{y0: y, y1: y + 1, x0: 1, x1: w - 1, focus: fi})
		}
		lines = append(lines, line)
		y++
	}

	addSection("After")
	for _, id := range []string{"sync", "tag-next", "push", "reinstall-local"} {
		renderStage(id, id, "")
	}

	addSection("Batch")
	preview := "would run: " + strings.Join(composeArgvFromRecipe(m.toRecipe()), " ")
	for len(preview) > 0 {
		chunk := preview
		if dashRuneLen(chunk) > innerW-1 {
			// cut by runes
			r := []rune(preview)
			chunk = string(r[:innerW-1])
			preview = string(r[innerW-1:])
		} else {
			preview = ""
		}
		inner := " " + dashPaint(m.color, ansiGrey, chunk)
		// pad
		plainLen := 1 + dashRuneLen(chunk)
		padN := innerW - plainLen
		if padN < 0 {
			padN = 0
		}
		line := paintBorder("│") + inner + strings.Repeat(" ", padN) + paintBorder("│")
		addPlain(line)
		y++
	}

	// RUN ALL + CANCEL on one line
	rai := focusIdx(focusRunAll, "run-all")
	cai := focusIdx(focusCancel, "cancel")
	rf := m.isFocused(focusRunAll, "run-all")
	cf := m.isFocused(focusCancel, "cancel")
	runAll := "[ RUN ALL ]"
	if m.loadingID == "run-all" {
		runAll = "[ " + m.spinnerGlyph() + " RUN… ]"
	} else if rf {
		runAll = "[▶ RUN ALL ]"
	}
	cancel := "[ CANCEL ]"
	if cf && m.loadingID == "" {
		cancel = "[▶ CANCEL ]"
	}
	// Split line: left half RUN ALL, right half CANCEL
	half := innerW / 2
	leftBtn := dashPad(" "+runAll, half)
	rightBtn := dashPad(cancel, innerW-half)
	var runAllVis, cancelVis string
	if m.loadingID == "run-all" {
		runAllVis = dashPaint(m.color, ansiBoldYellow, leftBtn)
	} else if rf {
		runAllVis = dashPaint(m.color, ansiBoldGreen, leftBtn)
	} else {
		runAllVis = dashPaint(m.color, ansiGreen, leftBtn)
	}
	if cf && m.loadingID == "" {
		cancelVis = dashPaint(m.color, ansiBoldYellow, rightBtn)
	} else {
		cancelVis = dashPaint(m.color, ansiGrey, rightBtn)
	}
	line := paintBorder("│") + runAllVis + cancelVis + paintBorder("│")
	if rai >= 0 {
		hits = append(hits, dashHit{y0: y, y1: y + 1, x0: 1, x1: 1 + half, focus: rai})
	}
	if cai >= 0 {
		hits = append(hits, dashHit{y0: y, y1: y + 1, x0: 1 + half, x1: w - 1, focus: cai})
	}
	lines = append(lines, line)
	y++

	footPlain := dashPad(" ↑↓ row · enter run · space toggle · click · q", innerW)
	addPlain(paintBorder("│") + dashPaint(m.color, ansiGrey, footPlain) + paintBorder("│"))
	y++
	addPlain(paintBorder(frameBottom(w)))
	_ = y

	m.hitmap = hits
	m.viewLines = len(lines)
	return strings.Join(lines, "\n") + "\n"
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// runDashboardTeaLoop runs a single Bubble Tea session until CANCEL.
// Stage / RUN ALL ops run in-process with a loading spinner (no tear-down flash).
func runDashboardTeaLoop(workDir string, ctx *invocationContext) error {
	_ = ctx // compose path uses nil ctx + skip via re-entry; events still recorded inside Run
	m := newTeaDashModel(workDir, "")
	// No alt-screen (inline UI). Mouse Y is terminal-absolute; mapMouseY
	// converts using height - viewLines (inline paint sits at bottom).
	p := tea.NewProgram(
		&m,
		tea.WithInput(os.Stdin),
		tea.WithOutput(os.Stdout),
		tea.WithMouseCellMotion(),
	)
	final, err := p.Run()
	if err != nil {
		return fmt.Errorf("wrk: dashboard tui: %w", err)
	}
	if _, ok := final.(*teaDashModel); !ok {
		return fmt.Errorf("wrk: unexpected tea model type %T", final)
	}
	return nil
}
