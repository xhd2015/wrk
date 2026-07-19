package tui

import (
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/xhd2015/dot-pkgs/go-pkgs/tui/mouse"
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

	// Inline mouse origin via shared tui/mouse package (CSI 6n + dual-origin).
	origin *mouse.Tracker
	cprCh  <-chan mouse.CPRMsg

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

// mapMouseY converts terminal-absolute mouse Y to view-local line index via
// ResolveMouseHit (known origin from CSI 6n when set, else dual-origin). Prefer
// the successful candidate's LocalY; if both miss, return a best-effort local.
func (m *teaDashModel) originYPtr() *int {
	if m.origin == nil {
		return nil
	}
	return m.origin.OriginY()
}

func (m *teaDashModel) mapMouseY(absY int) (local int, ok bool) {
	res := ResolveMouseHit(ResolveMouseHitOpts{
		AbsX:      0,
		AbsY:      absY,
		Height:    m.height,
		ViewLines: m.viewLines,
		OriginY:   m.originYPtr(),
		Hitmap:    dashHitsToHits(m.hitmap),
	})
	if res.OK {
		return res.LocalY, true
	}
	if oy := m.originYPtr(); oy != nil {
		local = absY - *oy
		if local >= 0 && (m.viewLines == 0 || local < m.viewLines) {
			return local, true
		}
	}
	if m.height > 0 && m.viewLines > 0 {
		origin := mouse.BottomOriginY(m.height, m.viewLines)
		local = absY - origin
		if local >= 0 && local < m.viewLines {
			return local, true
		}
	}
	if absY >= 0 && (m.viewLines == 0 || absY < m.viewLines) {
		return absY, true
	}
	return absY, false
}

func (m *teaDashModel) hitTest(x, y int) (h dashHit, ok bool) {
	hits := dashHitsToHits(m.hitmap)
	mh, ok := mouse.HitTest(hitsToMouse(hits), x, y)
	if !ok {
		return dashHit{}, false
	}
	got := mouseHitToDash(mh, hits)
	return dashHit{
		y0: got.Y0, y1: got.Y1,
		x0: got.X0, x1: got.X1,
		focus: got.Focus, runStage: got.RunStage,
	}, true
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
		opts:            opts,
		workDir:         workDir,
		status:          status,
		genCommitMsg:    true,
		commit:          true,
		agentRunner:     "commandcode",
		primaryDone:     false, // MERGE BACK default
		sync:            true,
		tagNext:         true,
		push:            true,
		reinstallLocal:  true,
		color:  dashColorOn(),
		origin: mouse.NewTracker(),
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

func (m *teaDashModel) Init() tea.Cmd {
	// Listen for CPR peeled from stdin (same path as mouse).
	return waitCPR(m.cprCh)
}

func (m *teaDashModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		if m.origin != nil {
			m.origin.OnResize()
			mouseDebugf("origin_invalidate", map[string]any{"layoutGen": m.origin.LayoutGen()})
		}
		return m, nil
	case mouse.CPRMsg:
		cmd := waitCPR(m.cprCh)
		ok := false
		if m.origin != nil {
			ok = m.origin.OnCPR(msg.Row1, msg.Col1)
		}
		mouseDebugf("cpr_raw", map[string]any{
			"row1": msg.Row1, "col1": msg.Col1,
			"ok": ok, "originY": originYVal(m.originYPtr()),
			"phase": fmt.Sprintf("%v", m.originPhase()),
		})
		if ok {
			mouseDebugf("cpr_ok", map[string]any{"originY": originYVal(m.originYPtr())})
		} else {
			mouseDebugf("cpr_fail_or_stale", map[string]any{
				"row1": msg.Row1, "phase": fmt.Sprintf("%v", m.originPhase()),
			})
		}
		return m, cmd
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
		// viewLines may change after op; View invalidates origin if needed.
		return m, nil
	case tea.MouseMsg:
		mouseDebugf("mouse_raw", map[string]any{
			"x": msg.X, "y": msg.Y,
			"action": fmt.Sprintf("%v", msg.Action),
			"button": fmt.Sprintf("%v", msg.Button),
			"type":   fmt.Sprintf("%v", msg.Type),
			"loadingID": m.loadingID,
			"height": m.height, "viewLines": m.viewLines,
			"originY": originYVal(m.originYPtr()),
			"addDisabled": m.addDisabled, "addAll": m.addAll,
			"width": m.width,
		})
		if m.loadingID != "" {
			mouseDebugf("mouse_ignore", map[string]any{"reason": "loading", "loadingID": m.loadingID})
			return m, nil // ignore clicks while running
		}
		if !dashMouseIsLeftClick(msg) {
			mouseDebugf("mouse_ignore", map[string]any{"reason": "not_left_press"})
			return m, nil
		}
		// Known origin from CSI 6n when available; else dual-origin.
		// If known-origin misses, fall back to dual-origin (stale/wrong CPR).
		hits := dashHitsToHits(m.hitmap)
		opts := ResolveMouseHitOpts{
			AbsX:      msg.X,
			AbsY:      msg.Y,
			Height:    m.height,
			ViewLines: m.viewLines,
			OriginY:   m.originYPtr(),
			Hitmap:    hits,
			Loading:   m.loadingID != "",
		}
		res := ResolveMouseHit(opts)
		mouseDebugf("mouse_resolve", map[string]any{
			"ok": res.OK, "localY": res.LocalY, "originKind": res.OriginKind,
			"hitRunStage": res.Hit.RunStage, "hitFocus": res.Hit.Focus,
			"hitY0": res.Hit.Y0, "hitY1": res.Hit.Y1, "hitX0": res.Hit.X0, "hitX1": res.Hit.X1,
			"originY": originYVal(m.originYPtr()),
			"hitmapN": len(hits),
			"hitmap": hitmapSummaryHits(hits),
		})
		if !res.OK && m.originYPtr() != nil {
			opts2 := opts
			opts2.OriginY = nil
			res2 := ResolveMouseHit(opts2)
			mouseDebugf("mouse_resolve_fallback_dual", map[string]any{
				"ok": res2.OK, "localY": res2.LocalY, "originKind": res2.OriginKind,
				"hitRunStage": res2.Hit.RunStage, "hitFocus": res2.Hit.Focus,
			})
			if res2.OK {
				res = res2
			}
		}
		if !res.OK {
			mouseDebugf("mouse_action", map[string]any{"action": "miss", "absY": msg.Y, "absX": msg.X})
			return m, nil
		}
		if res.Hit.RunStage != "" {
			mouseDebugf("mouse_action", map[string]any{"action": "startStageRun", "stage": res.Hit.RunStage})
			return m.startStageRun(res.Hit.RunStage)
		}
		if res.Hit.Focus >= 0 {
			mouseDebugf("mouse_action", map[string]any{"action": "activateFocus", "focus": res.Hit.Focus})
			m.cursor = res.Hit.Focus
			return m.activateFocus()
		}
		mouseDebugf("mouse_action", map[string]any{"action": "noop_hit"})
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
	disabled := m.stageRunDisabled(stageID)
	mouseDebugf("startStageRun", map[string]any{
		"stageID": stageID, "loadingID": m.loadingID,
		"disabled": disabled, "addDisabled": m.addDisabled,
	})
	if m.loadingID != "" || disabled {
		mouseDebugf("startStageRun_blocked", map[string]any{
			"stageID": stageID,
			"reason": map[string]any{
				"loading": m.loadingID != "",
				"disabled": disabled,
			},
		})
		return m, nil
	}
	m.loadingID = stageID
	m.loadingFrame = 0
	m.status = "running…"
	if stageID == "add-changes" {
		mouseDebugf("startStageRun_ok", map[string]any{"stageID": stageID, "path": "gitAddAll"})
		return m, tea.Batch(
			m.dashRunOpCmd(stageID, true, Recipe{}),
			dashSpinnerTick(),
		)
	}
	r, ok := m.singleStageRecipe(stageID)
	if !ok {
		mouseDebugf("startStageRun_blocked", map[string]any{"stageID": stageID, "reason": "no_recipe"})
		m.loadingID = ""
		return m, nil
	}
	mouseDebugf("startStageRun_ok", map[string]any{"stageID": stageID, "path": "compose"})
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

func (m *teaDashModel) originPhase() mouse.Phase {
	if m.origin == nil {
		return mouse.PhaseUnknown
	}
	return m.origin.Phase()
}

func (m *teaDashModel) View() string {
	s := m.renderView()
	if m.origin != nil {
		if suf := m.origin.FrameSuffix(m.height, m.viewLines); suf != "" {
			s += suf
			mouseDebugf("cpr_emit", map[string]any{
				"layoutGen": m.origin.LayoutGen(),
				"viewLines": m.viewLines, "height": m.height,
			})
		}
	}
	return s
}

// waitCPR delivers the next peeled CPR as mouse.CPRMsg and must be re-armed.
func waitCPR(ch <-chan mouse.CPRMsg) tea.Cmd {
	if ch == nil {
		return nil
	}
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return nil
		}
		return msg
	}
}
