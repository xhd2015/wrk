package tui

import (
	"strings"
)

// --- compact layout helpers ---

// dashLayoutWidth clamps the dashboard frame width.
// ~1.5× the historical 56–72 clamp so stage rows fit inline meta + [ Run ].
func dashLayoutWidth(termW int) int {
	w := termW
	if w < 84 {
		w = 84
	}
	if w > 108 {
		w = 108
	}
	return w
}

// stageInlineMeta returns in-row meta for the stage operation line (no leading separator).
// Only folds the former secondary line onto the same row as the stage — not a full redesign.
// Priority: msg > result > settled preview > pending > empty.
// green is true for msg/result; previews/pending stay gray.
func (m *teaDashModel) stageInlineMeta(id string) (meta string, green bool) {
	if id == "gen-commit-msg" {
		if subj := strings.TrimSpace(m.genSubject); subj != "" {
			return "msg: " + subj, true
		}
	}
	if m.stageResults != nil {
		res := strings.TrimSpace(m.stageResults[id])
		// Don't re-show subject as result under gen-commit-msg when msg: already preferred.
		if res != "" && !(id == "gen-commit-msg" && strings.TrimSpace(m.genSubject) != "") {
			return "result: " + res, true
		}
	}
	if !m.previewSettledFor(id) {
		return "preview: …", false
	}
	if m.previews != nil {
		if text := strings.TrimSpace(m.previews[id]); text != "" {
			return "preview: " + text, false
		}
	}
	return "", false
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

	// Section horizontal rules: ├─ Pre ──────────────┤
	// Always emitted for Pre / Main / After / Batch / Log (full-width HR, not plain text labels).
	addSection := func(name string) {
		if !m.color {
			addPlain(frameSection(w, name))
			y++
			return
		}
		// Color: grey box + rule, bold-cyan title so the HR stays readable.
		// Geometry matches frameSection: ├ + "─ Name " + ─*rest + ┤
		body := "─ " + name + " "
		rest := w - 2 - dashRuneLen(body)
		if rest < 0 {
			body = dashPad(body, w-2)
			rest = 0
		}
		line := dashPaint(true, ansiGrey, "├") +
			dashPaint(true, ansiGrey, "─") +
			" " + dashPaint(true, ansiBoldCyan, name) + " " +
			dashPaint(true, ansiGrey, strings.Repeat("─", rest)) +
			dashPaint(true, ansiGrey, "┤")
		addPlain(line)
		y++
	}

	// Compact stage-run marker (ok / ! / skip) for phase-aware pipeline feedback.
	stageMark := func(id string) (plain, painted string) {
		switch m.stageRunState(id) {
		case StageOK:
			plain = " ok"
			painted = dashPaint(m.color, ansiGreen, " ok")
		case StageError:
			plain = " !"
			painted = dashPaint(m.color, ansiRed, " !")
		case StageSkipped:
			plain = " ·"
			painted = dashPaint(m.color, ansiGrey, " ·")
		case StageQueued:
			plain = " …"
			painted = dashPaint(m.color, ansiGrey, " …")
		default:
			return "", ""
		}
		return plain, painted
	}

	// Stage row: one line (title + inline meta + Run chip). No secondary line.
	renderStage := func(id, title, extra string) {
		ti := focusIdx(focusStage, id)
		rowF := m.isStageRowFocused(id)
		g := m.glyph(id)
		st := m.stageRunState(id)
		loadingHere := m.loadingID == id || st == StageRunning

		markPlain, markPaint := stageMark(id)
		meta, metaGreen := m.stageInlineMeta(id)
		metaPlain := ""
		if meta != "" {
			metaPlain = " · " + meta
		}

		left := g + " " + title
		if extra != "" {
			left += "  " + extra
		}
		left += markPlain + metaPlain
		if rowF {
			left = "▶" + left
		} else {
			left = " " + left
		}
		// Truncate whole left column so Run chip stays fixed width.
		leftTrunc := dashRuneLen(left) > leftW
		if leftTrunc {
			left = dashPad(left, leftW)
		}

		runLabel := "[ Run ]"
		if m.stageRunDisabled(id) {
			runLabel = "[  —  ]"
		} else if loadingHere {
			runLabel = "[ " + m.spinnerGlyph() + " ]"
		}

		runPlain := strings.Repeat(" ", max(0, runW-dashRuneLen(runLabel))) + runLabel
		if dashRuneLen(runPlain) > runW {
			runPlain = dashPad(runLabel, runW)
		}

		var leftVis string
		if leftTrunc {
			leftVis = left
			if m.color {
				leftVis = dashPaint(true, ansiGrey, left)
			}
		} else {
			gCol := m.paintGlyph(g)
			titlePart := title
			extraPart := extra
			if m.color {
				if g == dashGlyphDisabled || (id == "add-changes" && m.addDisabled) || st == StageSkipped {
					titlePart = dashPaint(true, ansiGrey, title)
					extraPart = dashPaint(true, ansiGrey, extra)
				} else {
					extraPart = dashPaint(true, ansiGrey, extra)
				}
				if rowF && !loadingHere {
					titlePart = dashPaint(true, ansiBoldGreen, title)
				}
			}
			leftVis = " "
			if rowF {
				leftVis = dashPaint(m.color, ansiGreen, "▶")
			}
			leftVis += gCol + " " + titlePart
			if extra != "" {
				leftVis += "  " + extraPart
			}
			leftVis += markPaint
			if meta != "" {
				metaCol := ansiGrey
				if metaGreen {
					metaCol = ansiGreen
				}
				leftVis += dashPaint(m.color, metaCol, " · "+meta)
			}
			padN := leftW - dashRuneLen(left)
			if padN < 0 {
				padN = 0
			}
			leftVis += strings.Repeat(" ", padN)
		}

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
		st := m.stageRunState(id)
		markPlain, markPaint := stageMark(id)
		meta, metaGreen := m.stageInlineMeta(id)
		metaPlain := ""
		if meta != "" {
			metaPlain = " · " + meta
		}
		// Spinner hint while main stage is active during RUN ALL.
		runHint := ""
		if st == StageRunning {
			runHint = " " + m.spinnerGlyph()
		}
		prefix := " "
		if foc {
			prefix = "▶"
		}
		body := prefix + g + " " + name + "   " + desc + markPlain + metaPlain + runHint
		bodyTrunc := dashRuneLen(body) > innerW
		if bodyTrunc {
			body = dashPad(body, innerW)
		}

		var vis string
		if bodyTrunc {
			vis = body
			if m.color {
				vis = dashPaint(true, ansiGrey, body)
			}
		} else {
			gCol := m.paintGlyph(g)
			nameCol := name
			descCol := dashPaint(m.color, ansiGrey, desc)
			if m.mainDisabled || st == StageSkipped {
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
			vis = prefix + gCol + " " + nameCol + "   " + descCol + markPaint
			if meta != "" {
				metaCol := ansiGrey
				if metaGreen {
					metaCol = ansiGreen
				}
				vis += dashPaint(m.color, metaCol, " · "+meta)
			}
			if runHint != "" {
				vis += dashPaint(m.color, ansiBoldYellow, runHint)
			}
			padN := innerW - dashRuneLen(body)
			if padN < 0 {
				padN = 0
			}
			vis += strings.Repeat(" ", padN)
		}
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
	previewArgv := []string{}
	if m.opts.ComposeArgv != nil {
		previewArgv = m.opts.ComposeArgv(m.toRecipe())
	}
	preview := "would run: " + strings.Join(previewArgv, " ")
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

	// Log section: fixed viewport of dashLogViewLines body rows (always reserved).
	// Empty slots are blank pads so new logs never change viewLines / mouse origin.
	// Shows last N logs (newest at bottom); empty ring: "(no log yet)" + blanks.
	addSection("Log")
	addLogBodyLine := func(text string) {
		// Leading space + truncated text, pad to innerW (blank text = empty reserved row).
		maxText := innerW - 1
		if maxText < 1 {
			maxText = 1
		}
		r := []rune(text)
		if len(r) > maxText {
			text = string(r[:maxText])
		}
		plainLen := 1 + dashRuneLen(text)
		padN := innerW - plainLen
		if padN < 0 {
			padN = 0
		}
		painted := ""
		if text != "" {
			painted = dashPaint(m.color, ansiGrey, text)
		}
		line := paintBorder("│") + " " + painted + strings.Repeat(" ", padN) + paintBorder("│")
		// No hitmap entries — log lines are not focus targets.
		addPlain(line)
		y++
	}
	// Build exactly dashLogViewLines content rows (pad at top).
	viewport := make([]string, 0, dashLogViewLines)
	if len(m.logs) == 0 {
		viewport = append(viewport, "(no log yet)")
	} else {
		logStart := 0
		if len(m.logs) > dashLogViewLines {
			logStart = len(m.logs) - dashLogViewLines
		}
		for _, ll := range m.logs[logStart:] {
			text := ll.Text
			if ll.Stage != "" {
				text = ll.Stage + " │ " + ll.Text
			}
			viewport = append(viewport, text)
		}
	}
	// Pad at top so viewport is always dashLogViewLines tall.
	for len(viewport) < dashLogViewLines {
		viewport = append([]string{""}, viewport...)
	}
	if len(viewport) > dashLogViewLines {
		viewport = viewport[len(viewport)-dashLogViewLines:]
	}
	for _, text := range viewport {
		addLogBodyLine(text)
	}

	footPlain := dashPad(" ↑↓ row · enter run · space toggle · click · q", innerW)
	addPlain(paintBorder("│") + dashPaint(m.color, ansiGrey, footPlain) + paintBorder("│"))
	y++
	addPlain(paintBorder(frameBottom(w)))
	_ = y

	m.hitmap = hits
	// viewLines must match Bubble Tea's rendered line count for bottom-origin
	// mouse mapping. A trailing "\n" makes strings.Split produce an extra empty
	// segment → linesRendered = len(lines)+1 → origin off-by-one (clicks hit the
	// row above: gen-commit-msg Run → add-changes).
	m.viewLines = len(lines)
	return strings.Join(lines, "\n")
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}
