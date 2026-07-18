package tui

import (
	"io"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

// Dashboard glyphs (fine-grained; never [x]/[X]).
const (
	dashGlyphOn       = "[•]"
	dashGlyphOff      = "[ ]"
	dashGlyphDisabled = "[-]"
)

// Hermetic dry-run env (same name as wrkcli; do not import parent package).
const envDashboardDryRun = "WRK_DASHBOARD_DRY_RUN"

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
	y0, y1   int
	x0, x1   int // if x1<=x0, whole line matches
	focus    int
	runStage string // single-phase run id; exclusive with normal activate when set
}

// teaDashModel is the Bubble Tea model for bare `wrk` on a real TTY.
type teaDashModel struct {
	opts RunDashboardOpts

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

	focus     []dashFocus
	cursor    int
	width     int
	height    int // terminal rows (WindowSizeMsg); for mouse Y mapping without alt-screen
	hitmap    []dashHit
	viewLines int // last rendered line count (hitmap local Y is 0..viewLines-1)
	color     bool

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
	ansiCyan       = "\x1b[36m"
	ansiBoldGreen  = "\x1b[1;32m"
	ansiBoldCyan   = "\x1b[1;36m"
	ansiBoldYellow = "\x1b[1;33m"
	ansiYellow     = "\x1b[33m"
	ansiRed        = "\x1b[31m"
	ansiGreen      = "\x1b[32m"
	ansiGrey       = "\x1b[90m"
	ansiReset      = "\x1b[0m"
)

func dashColorOn() bool {
	return os.Getenv("NO_COLOR") == ""
}

func (m *teaDashModel) hasAddableDirt() bool {
	if m.opts.HasAddableDirt == nil {
		return false
	}
	return m.opts.HasAddableDirt(m.workDir)
}

func (m *teaDashModel) isMainCheckout() bool {
	if m.opts.IsMainCheckout == nil {
		return true
	}
	return m.opts.IsMainCheckout(m.workDir)
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

func newTeaDashModel(opts RunDashboardOpts) teaDashModel {
	workDir := opts.WorkDir
	status := opts.Status
	m := teaDashModel{
		opts:           opts,
		workDir:        workDir,
		status:         status,
		genCommitMsg:   true,
		commit:         true,
		agentRunner:    "commandcode",
		primaryDone:    false, // MERGE BACK default
		sync:           true,
		tagNext:        true,
		push:           true,
		reinstallLocal: true,
		color:          dashColorOn(),
	}
	addable := m.hasAddableDirt()
	onMain := m.isMainCheckout()
	m.addAll = addable
	m.addDisabled = !addable
	m.mainDisabled = onMain
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
		m.addDisabled = !m.hasAddableDirt()
		if m.addDisabled {
			m.addAll = false
		}
		m.mainDisabled = m.isMainCheckout()
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
func (m *teaDashModel) dashRunOpCmd(loadingID string, addOnly bool, recipe Recipe) tea.Cmd {
	workDir := m.workDir
	opts := m.opts
	return func() tea.Msg {
		status, err := runDashboardOpCaptured(opts, workDir, addOnly, recipe)
		return dashOpDoneMsg{loadingID: loadingID, status: status, err: err}
	}
}

// runDashboardOpCaptured executes git-add or compose with stdio redirected to a pipe.
func runDashboardOpCaptured(opts RunDashboardOpts, workDir string, addOnly bool, recipe Recipe) (status string, err error) {
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
		if opts.GitAddAll != nil {
			runErr = opts.GitAddAll(workDir)
		}
		if runErr == nil {
			status = "ok  staged: git add -A"
		}
	} else {
		if opts.RunCompose != nil {
			runErr = opts.RunCompose(workDir, recipe)
		}
		if runErr == nil {
			argvParts := []string{}
			if opts.ComposeArgv != nil {
				argvParts = opts.ComposeArgv(recipe)
			}
			argv := strings.Join(argvParts, " ")
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
			m.dashRunOpCmd(stageID, true, Recipe{}),
			dashSpinnerTick(),
		)
	}
	r, ok := m.singleStageRecipe(stageID)
	if !ok {
		m.loadingID = ""
		return m, nil
	}
	return m, tea.Batch(
		m.dashRunOpCmd(stageID, false, r),
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
		m.dashRunOpCmd("run-all", false, r),
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

func (m *teaDashModel) singleStageRecipe(id string) (Recipe, bool) {
	r := Recipe{AgentRunner: m.agentRunner, DryRun: os.Getenv(envDashboardDryRun) == "1"}
	switch id {
	case "gen-commit-msg":
		r.GenCommitMsg = true
		r.AddAll = m.addAll && !m.addDisabled
		return r, true
	case "commit":
		r.GenCommitMsg = true
		r.AddAll = m.addAll && !m.addDisabled
		r.Commit = true
		return r, true
	case "sync":
		r.Sync = true
		return r, true
	case "tag-next":
		r.TagNext = true
		return r, true
	case "push":
		r.Push = true
		return r, true
	case "reinstall-local":
		r.ReinstallLocal = true
		return r, true
	default:
		return r, false
	}
}

func (m *teaDashModel) toRecipe() Recipe {
	return Recipe{
		AddAll:         m.addAll && !m.addDisabled,
		GenCommitMsg:   m.genCommitMsg,
		Commit:         m.commit,
		AgentRunner:    m.agentRunner,
		Done:           m.primaryDone && !m.mainDisabled,
		MergeBack:      !m.primaryDone && !m.mainDisabled,
		Sync:           m.sync,
		TagNext:        m.tagNext,
		Push:           m.push,
		ReinstallLocal: m.reinstallLocal,
		DryRun:         os.Getenv(envDashboardDryRun) == "1",
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
