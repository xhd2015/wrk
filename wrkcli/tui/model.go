package tui

import (
	"bufio"
	"fmt"
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

// LogLine is one captured stdout/stderr line from a stage / RUN ALL op.
type LogLine struct {
	Stage string
	Level string // optional; empty for plain process output
	Text  string
	At    time.Time
}

const (
	maxDashLogs      = 200 // ring buffer capacity
	dashLogViewLines = 3   // fixed Log viewport height (always reserved; no layout bounce)
)

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

	// Streaming op logs (ring buffer + live channel while an op runs).
	logs  []LogLine
	logCh <-chan dashLogLineMsg // set while op streams; re-armed via listenLogs

	// Phase-aware pipeline events (RUN ALL / optional single-stage).
	stageRun map[string]StageRunState
	eventCh  <-chan dashStageEventMsg // set while pipeline streams; re-armed via listenEvents

	// Optional one-line stage previews filled asynchronously after first paint.
	// Keyed by stage id: add-changes, gen-commit-msg, commit, merge-back, done,
	// sync, tag-next, push, reinstall-local.
	// Secondary lines are always reserved (stable viewLines); empty settled → blank.
	previews map[string]string
	// previewSettled[id]==true once a dashPreviewMsg has been applied for id.
	// Unsettled stages show "…" in the reserved secondary slot (no layout bounce).
	previewSettled map[string]bool

	// Brief per-stage results from StageEvent.Result after a successful run.
	// Prefer over preview for that stage until the next run clears them.
	stageResults map[string]string
	// Structured gen-commit message (subject line + optional body).
	genSubject string
	genBody    string

	quitOutcome int // only cancel leaves the program
}

// dashStageIDs is the ordered list of stages that may show a preview line.
var dashStageIDs = []string{
	"add-changes", "gen-commit-msg", "commit",
	"merge-back", "done",
	"sync", "tag-next", "push", "reinstall-local",
}

// previewJobTimeout is a soft cap per stage preview so expensive helpers cannot stall.
const previewJobTimeout = 1500 * time.Millisecond

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

// dashLogLineMsg is one streamed stdout/stderr line from a running op.
type dashLogLineMsg struct {
	Stage string
	Level string
	Text  string
}

// dashLogClosedMsg signals the log channel was closed (op finished streaming).
type dashLogClosedMsg struct{}

// dashStageEventMsg is one stage transition from a phase-aware pipeline run.
type dashStageEventMsg StageEvent

// dashEventsClosedMsg signals the stage-event channel was closed.
type dashEventsClosedMsg struct{}

// dashTickMsg advances the loading spinner while an op runs.
type dashTickMsg struct{}

// dashPreviewMsg delivers one stage's async one-line preview text and any
// captured diagnostics for the Log panel (normal log lines, not tty prints).
type dashPreviewMsg struct {
	StageID string
	Text    string
	Logs    []string // e.g. git stderr lines; append to log ring as normal logs
}

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
		origin:         mouse.NewTracker(),
		previews:       make(map[string]string),
		previewSettled: make(map[string]bool),
		stageResults:   make(map[string]string),
		stageRun:       make(map[string]StageRunState),
	}
	addable := m.hasAddableDirt()
	onMain := m.isMainCheckout()
	m.addAll = addable
	m.addDisabled = !addable
	m.mainDisabled = onMain
	// No StagePreview injector: settle all empty so slots are blank (not forever "…").
	// With injector, Init/previewCmds marks pending until async msgs settle.
	if m.opts.StagePreview == nil {
		m.markPreviewsSettledEmpty()
	} else {
		m.markPreviewsPending()
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
	// CPR listener + async stage previews (do not block first paint).
	return tea.Batch(waitCPR(m.cprCh), m.previewCmds())
}

// markPreviewsPending marks every stage secondary slot as loading ("…").
// Used on Init (with StagePreview) and after ops before re-fetching previews.
func (m *teaDashModel) markPreviewsPending() {
	if m.previewSettled == nil {
		m.previewSettled = make(map[string]bool)
	}
	for _, id := range dashStageIDs {
		m.previewSettled[id] = false
	}
}

// markPreviewsSettledEmpty marks all stages settled with no preview text (blank slots).
func (m *teaDashModel) markPreviewsSettledEmpty() {
	if m.previewSettled == nil {
		m.previewSettled = make(map[string]bool)
	}
	if m.previews == nil {
		m.previews = make(map[string]string)
	}
	for _, id := range dashStageIDs {
		m.previewSettled[id] = true
		delete(m.previews, id)
	}
}

// previewSettledFor reports whether the stage's async preview has been applied.
func (m *teaDashModel) previewSettledFor(id string) bool {
	if m.previewSettled == nil {
		return false
	}
	return m.previewSettled[id]
}

// previewCmds starts one background tea.Cmd per stage. Each soft-fails to empty
// preview text on panic/timeout; results arrive as dashPreviewMsg and never fail the UI.
// Captured diagnostics are carried in Logs for the Log panel (not discarded, not tty).
// Re-marks all stages pending so the reserved slot shows "…" until each settles.
func (m *teaDashModel) previewCmds() tea.Cmd {
	if m.opts.StagePreview == nil {
		m.markPreviewsSettledEmpty()
		return nil
	}
	m.markPreviewsPending()
	workDir := m.workDir
	fn := m.opts.StagePreview
	cmds := make([]tea.Cmd, 0, len(dashStageIDs))
	for _, id := range dashStageIDs {
		stageID := id
		cmds = append(cmds, func() tea.Msg {
			res := runStagePreview(fn, workDir, stageID)
			return dashPreviewMsg{
				StageID: stageID,
				Text:    res.Preview,
				Logs:    res.Logs,
			}
		})
	}
	return tea.Batch(cmds...)
}

// runStagePreview invokes StagePreview with a short timeout and recovers panics.
func runStagePreview(fn func(workDir, stageID string) StagePreviewResult, workDir, stageID string) StagePreviewResult {
	if fn == nil {
		return StagePreviewResult{}
	}
	type result struct{ r StagePreviewResult }
	ch := make(chan result, 1)
	go func() {
		defer func() {
			if recover() != nil {
				ch <- result{StagePreviewResult{}}
			}
		}()
		r := fn(workDir, stageID)
		r.Preview = strings.TrimSpace(r.Preview)
		ch <- result{r}
	}()
	select {
	case r := <-ch:
		return r.r
	case <-time.After(previewJobTimeout):
		return StagePreviewResult{}
	}
}

// applyPreview stores or clears one stage preview and marks the slot settled.
// Empty text deletes the key (settled blank slot); non-empty stores truncated text.
func (m *teaDashModel) applyPreview(stageID, text string) {
	stageID = strings.TrimSpace(stageID)
	if stageID == "" {
		return
	}
	if m.previews == nil {
		m.previews = make(map[string]string)
	}
	if m.previewSettled == nil {
		m.previewSettled = make(map[string]bool)
	}
	text = strings.TrimSpace(text)
	if text == "" {
		delete(m.previews, stageID)
		m.previewSettled[stageID] = true
		return
	}
	// Keep one short line for layout stability.
	if len(text) > 60 {
		text = text[:57] + "..."
	}
	m.previews[stageID] = text
	m.previewSettled[stageID] = true
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
	case dashLogLineMsg:
		m.appendLog(LogLine{
			Stage: msg.Stage,
			Level: msg.Level,
			Text:  msg.Text,
			At:    time.Now(),
		})
		// Re-arm listener while the channel is still open.
		if m.logCh != nil {
			return m, listenLogs(m.logCh)
		}
		return m, nil
	case dashLogClosedMsg:
		m.logCh = nil
		return m, nil
	case dashStageEventMsg:
		m.applyStageEvent(StageEvent(msg))
		if m.eventCh != nil {
			return m, listenEvents(m.eventCh)
		}
		return m, nil
	case dashEventsClosedMsg:
		m.eventCh = nil
		return m, nil
	case dashPreviewMsg:
		m.applyPreview(msg.StageID, msg.Text)
		// Captured diagnostics (e.g. git stderr) are normal Log lines — never discarded.
		for _, line := range msg.Logs {
			line = strings.TrimSpace(line)
			if line == "" {
				continue
			}
			m.appendLog(LogLine{
				Stage: msg.StageID,
				Level: "info",
				Text:  line,
				At:    time.Now(),
			})
		}
		return m, nil
	case dashOpDoneMsg:
		m.loadingID = ""
		m.loadingFrame = 0
		// Finalize stages that never got a terminal event (fallback RunCompose, race).
		for id, st := range m.stageRun {
			switch st {
			case StageRunning:
				if msg.err != nil {
					m.stageRun[id] = StageError
				} else {
					m.stageRun[id] = StageOK
				}
			case StageQueued:
				// On error remaining queued stages were not reached → skipped.
				// On success without per-stage events (fallback) → ok.
				if msg.err != nil {
					m.stageRun[id] = StageSkipped
				} else {
					m.stageRun[id] = StageOK
				}
			}
		}
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
		// logCh / eventCh cleared by closed msgs after channels close.
		// Light preview refresh after ops (cheap git helpers only).
		return m, m.previewCmds()
	case tea.MouseMsg:
		mouseDebugf("mouse_raw", map[string]any{
			"x": msg.X, "y": msg.Y,
			"action":    fmt.Sprintf("%v", msg.Action),
			"button":    fmt.Sprintf("%v", msg.Button),
			"type":      fmt.Sprintf("%v", msg.Type),
			"loadingID": m.loadingID,
			"height":    m.height, "viewLines": m.viewLines,
			"originY":     originYVal(m.originYPtr()),
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
			"hitmap":  hitmapSummaryHits(hits),
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

// appendLog adds a line to the ring buffer, dropping oldest entries past maxDashLogs.
func (m *teaDashModel) appendLog(line LogLine) {
	m.logs = append(m.logs, line)
	if len(m.logs) > maxDashLogs {
		// Truncate from head; copy to avoid unbounded underlying array growth.
		n := len(m.logs) - maxDashLogs
		m.logs = append([]LogLine(nil), m.logs[n:]...)
	}
}

// listenLogs waits for the next streamed log line (or channel close).
func listenLogs(ch <-chan dashLogLineMsg) tea.Cmd {
	if ch == nil {
		return nil
	}
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return dashLogClosedMsg{}
		}
		return msg
	}
}

// listenEvents waits for the next stage event (or channel close).
func listenEvents(ch <-chan dashStageEventMsg) tea.Cmd {
	if ch == nil {
		return nil
	}
	return func() tea.Msg {
		msg, ok := <-ch
		if !ok {
			return dashEventsClosedMsg{}
		}
		return msg
	}
}

// applyStageEvent updates stageRun and status from a pipeline StageEvent.
// On ok, stores Result / gen Subject+Body for the view and optional Log body lines.
func (m *teaDashModel) applyStageEvent(ev StageEvent) {
	if m.stageRun == nil {
		m.stageRun = make(map[string]StageRunState)
	}
	id := strings.TrimSpace(ev.StageID)
	if id == "" {
		return
	}
	switch strings.ToLower(strings.TrimSpace(ev.Kind)) {
	case "start":
		m.stageRun[id] = StageRunning
		m.status = "running  " + id + "…"
	case "ok":
		m.stageRun[id] = StageOK
		if res := strings.TrimSpace(ev.Result); res != "" {
			if m.stageResults == nil {
				m.stageResults = make(map[string]string)
			}
			m.stageResults[id] = res
		}
		if subj := strings.TrimSpace(ev.Subject); subj != "" {
			m.genSubject = subj
			// Keep a stage result for gen-commit-msg even when Result was empty
			// so other consumers can read stageResults["gen-commit-msg"].
			if id == "gen-commit-msg" || id == "commit" {
				if m.stageResults == nil {
					m.stageResults = make(map[string]string)
				}
				if strings.TrimSpace(m.stageResults["gen-commit-msg"]) == "" {
					m.stageResults["gen-commit-msg"] = subj
				}
			}
		}
		if body := strings.TrimSpace(ev.Body); body != "" {
			m.genBody = body
			// First body lines go to the Log section for detail without crowding the stage row.
			for _, line := range strings.Split(body, "\n") {
				line = strings.TrimSpace(line)
				if line == "" {
					continue
				}
				m.appendLog(LogLine{Stage: id, Text: line, At: time.Now()})
			}
		}
	case "error":
		m.stageRun[id] = StageError
		if strings.TrimSpace(ev.Err) != "" {
			m.status = "error: " + ev.Err
		} else {
			m.status = "error: " + id
		}
	case "skipped":
		m.stageRun[id] = StageSkipped
	}
}

// resetStageRunForPlan sets planned stages to queued and others idle.
// Clears results for stages in the plan so stale msg/result lines do not linger.
func (m *teaDashModel) resetStageRunForPlan(plan []string) {
	m.stageRun = make(map[string]StageRunState)
	for _, id := range dashStageIDs {
		m.stageRun[id] = StageIdle
	}
	clearGen := false
	for _, id := range plan {
		m.stageRun[id] = StageQueued
		if m.stageResults != nil {
			delete(m.stageResults, id)
		}
		if id == "gen-commit-msg" || id == "commit" {
			clearGen = true
		}
	}
	if clearGen {
		m.genSubject = ""
		m.genBody = ""
	}
}

// SplitCommitMessage splits a full commit message into subject (first line) and body.
// Pure helper for tests and any caller that holds a multi-line message string.
func SplitCommitMessage(msg string) (subject, body string) {
	msg = strings.ReplaceAll(msg, "\r\n", "\n")
	msg = strings.TrimRight(msg, "\n")
	if msg == "" {
		return "", ""
	}
	i := strings.IndexByte(msg, '\n')
	if i < 0 {
		return strings.TrimSpace(msg), ""
	}
	subject = strings.TrimSpace(msg[:i])
	body = strings.TrimSpace(msg[i+1:])
	return subject, body
}

// stageRunState returns the current run state for id (idle if unset).
func (m *teaDashModel) stageRunState(id string) StageRunState {
	if m.stageRun == nil {
		return StageIdle
	}
	if st, ok := m.stageRun[id]; ok {
		return st
	}
	return StageIdle
}

// dashRunOpCmd runs a dashboard op off the UI thread.
// User-visible process output is fed into logCh (Log panel), never left on the real tty.
// When usePipeline is true and opts.RunPipeline is set, stages emit on eventCh and may
// call LogFunc which also writes to logCh (TUI-safe structured logs).
// Caller must tea.Batch this with listenLogs / listenEvents. Channels are closed on finish.
func (m *teaDashModel) dashRunOpCmd(loadingID string, addOnly bool, recipe Recipe, usePipeline bool, logCh chan<- dashLogLineMsg, eventCh chan<- dashStageEventMsg) tea.Cmd {
	workDir := m.workDir
	opts := m.opts
	return func() tea.Msg {
		status, err := runDashboardOpCaptured(opts, workDir, addOnly, recipe, loadingID, usePipeline, logCh, eventCh)
		if logCh != nil {
			close(logCh)
		}
		if eventCh != nil {
			close(eventCh)
		}
		return dashOpDoneMsg{loadingID: loadingID, status: status, err: err}
	}
}

// tuiLogFunc returns a LogFunc that posts lines to logCh for the Log panel.
// Never writes to the real stdout/stderr.
func tuiLogFunc(logCh chan<- dashLogLineMsg) LogFunc {
	return func(stage, line string) {
		if logCh == nil {
			return
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			return
		}
		logCh <- dashLogLineMsg{Stage: stage, Text: line}
	}
}

// runDashboardOpCaptured executes git-add, pipeline, or compose.
//
// TUI log design:
//   - Structured pipeline notes go through LogFunc → logCh → model → View Log section.
//   - Legacy bridge: for RunCompose / library code that still fmt-prints, process
//     stdout/stderr are temporarily redirected to a pipe and each line is mirrored
//     onto logCh. That is not the preferred API — prefer LogFunc from RunPipeline.
//   - Do not fmt.Print / log.Printf to the real tty while tea owns the screen.
func runDashboardOpCaptured(opts RunDashboardOpts, workDir string, addOnly bool, recipe Recipe, stage string, usePipeline bool, logCh chan<- dashLogLineMsg, eventCh chan<- dashStageEventMsg) (status string, err error) {
	logf := tuiLogFunc(logCh)

	// Legacy bridge: capture fmt/cli output from compose/git helpers that still
	// write to process stdout/stderr. Lines are re-emitted on logCh only.
	oldOut, oldErr := os.Stdout, os.Stderr
	r, w, pipeErr := os.Pipe()
	if pipeErr != nil {
		return "", pipeErr
	}
	os.Stdout, os.Stderr = w, w
	done := make(chan struct{})
	go func() {
		defer close(done)
		sc := bufio.NewScanner(r)
		// Allow long compose lines (default 64K is usually enough; bump for safety).
		sc.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for sc.Scan() {
			logf(stage, sc.Text())
		}
		_ = sc.Err()
	}()

	runErr := error(nil)
	if addOnly {
		if eventCh != nil {
			eventCh <- dashStageEventMsg{StageID: stage, Kind: "start"}
		}
		logf(stage, "git add -A")
		if opts.GitAddAll != nil {
			runErr = opts.GitAddAll(workDir)
		}
		if eventCh != nil {
			if runErr != nil {
				eventCh <- dashStageEventMsg{StageID: stage, Kind: "error", Err: runErr.Error()}
				logf(stage, "error: "+runErr.Error())
			} else {
				eventCh <- dashStageEventMsg{StageID: stage, Kind: "ok", Result: "staged"}
				logf(stage, "ok: staged")
			}
		}
		if runErr == nil {
			status = "ok  staged: git add -A"
		}
	} else if usePipeline && opts.RunPipeline != nil {
		emit := func(ev StageEvent) {
			if eventCh == nil {
				return
			}
			eventCh <- dashStageEventMsg(ev)
		}
		// Pipeline-structured logs use the same logCh (not the real tty).
		// Tag default stage as loadingID; pipeline may override per call.
		runErr = opts.RunPipeline(workDir, recipe, emit, logf)
		if runErr == nil {
			argvParts := []string{}
			if opts.ComposeArgv != nil {
				argvParts = opts.ComposeArgv(recipe)
			}
			argv := strings.Join(argvParts, " ")
			status = "ok  " + argv
		} else {
			logf(stage, "error: "+runErr.Error())
		}
	} else {
		if eventCh != nil && stage != "" && stage != "run-all" {
			eventCh <- dashStageEventMsg{StageID: stage, Kind: "start"}
		}
		logf(stage, "compose…")
		if opts.RunCompose != nil {
			runErr = opts.RunCompose(workDir, recipe)
		}
		if eventCh != nil && stage != "" && stage != "run-all" {
			if runErr != nil {
				eventCh <- dashStageEventMsg{StageID: stage, Kind: "error", Err: runErr.Error()}
				logf(stage, "error: "+runErr.Error())
			} else {
				eventCh <- dashStageEventMsg{StageID: stage, Kind: "ok"}
				logf(stage, "ok")
			}
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
				"loading":  m.loadingID != "",
				"disabled": disabled,
			},
		})
		return m, nil
	}
	m.loadingID = stageID
	m.loadingFrame = 0
	m.status = "running  " + stageID + "…"
	m.resetStageRunForPlan([]string{stageID})
	m.stageRun[stageID] = StageRunning
	if stageID == "add-changes" {
		mouseDebugf("startStageRun_ok", map[string]any{"stageID": stageID, "path": "gitAddAll"})
		logCh := make(chan dashLogLineMsg, 64)
		eventCh := make(chan dashStageEventMsg, 16)
		m.logCh = logCh
		m.eventCh = eventCh
		return m, tea.Batch(
			m.dashRunOpCmd(stageID, true, Recipe{}, false, logCh, eventCh),
			listenLogs(logCh),
			listenEvents(eventCh),
			dashSpinnerTick(),
		)
	}
	r, ok := m.singleStageRecipe(stageID)
	if !ok {
		mouseDebugf("startStageRun_blocked", map[string]any{"stageID": stageID, "reason": "no_recipe"})
		m.loadingID = ""
		m.stageRun[stageID] = StageIdle
		return m, nil
	}
	// Prefer phase-aware pipeline when injected so gen/commit get structured Subject/Result.
	usePipeline := m.opts.RunPipeline != nil
	mouseDebugf("startStageRun_ok", map[string]any{"stageID": stageID, "path": "compose", "usePipeline": usePipeline})
	logCh := make(chan dashLogLineMsg, 64)
	eventCh := make(chan dashStageEventMsg, 16)
	m.logCh = logCh
	m.eventCh = eventCh
	return m, tea.Batch(
		m.dashRunOpCmd(stageID, false, r, usePipeline, logCh, eventCh),
		listenLogs(logCh),
		listenEvents(eventCh),
		dashSpinnerTick(),
	)
}

func (m *teaDashModel) startRunAll() (tea.Model, tea.Cmd) {
	if m.loadingID != "" {
		return m, nil
	}
	r := m.toRecipe()
	plan := PlanRecipeStages(r)
	m.resetStageRunForPlan(plan)
	m.loadingID = "run-all"
	m.loadingFrame = 0
	m.status = "running…"
	logCh := make(chan dashLogLineMsg, 64)
	m.logCh = logCh
	usePipeline := m.opts.RunPipeline != nil
	var eventCh chan dashStageEventMsg
	if usePipeline {
		eventCh = make(chan dashStageEventMsg, 64)
		m.eventCh = eventCh
	}
	cmds := []tea.Cmd{
		m.dashRunOpCmd("run-all", false, r, usePipeline, logCh, eventCh),
		listenLogs(logCh),
		dashSpinnerTick(),
	}
	if eventCh != nil {
		cmds = append(cmds, listenEvents(eventCh))
	}
	return m, tea.Batch(cmds...)
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
