package tui

import (
	"encoding/json"
	"fmt"
	"os"
	"sync"
	"time"
)

// Mouse debug log path. Enabled when WRK_TUI_MOUSE_DEBUG is non-empty
// (any value except "0"/"false"). Appends JSON lines; never writes to the TTY.
//
//	export WRK_TUI_MOUSE_DEBUG=1
//	: > /tmp/wrk-tui-mouse.log   # optional clear
//	wrk
//
// Then: cat /tmp/wrk-tui-mouse.log
const mouseDebugLogPath = "/tmp/wrk-tui-mouse.log"

var mouseDebugMu sync.Mutex

func mouseDebugEnabled() bool {
	v := os.Getenv("WRK_TUI_MOUSE_DEBUG")
	if v == "" || v == "0" || v == "false" || v == "FALSE" {
		return false
	}
	return true
}

// mouseDebugf appends one JSON line to mouseDebugLogPath (best-effort, no panic).
func mouseDebugf(event string, fields map[string]any) {
	if !mouseDebugEnabled() {
		return
	}
	if fields == nil {
		fields = map[string]any{}
	}
	fields["ts"] = time.Now().Format(time.RFC3339Nano)
	fields["event"] = event
	b, err := json.Marshal(fields)
	if err != nil {
		return
	}
	mouseDebugMu.Lock()
	defer mouseDebugMu.Unlock()
	f, err := os.OpenFile(mouseDebugLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	_, _ = f.Write(append(b, '\n'))
	_ = f.Close()
}

func originYVal(p *int) any {
	if p == nil {
		return nil
	}
	return *p
}

func hitmapSummary(hits []dashHit) []map[string]any {
	out := make([]map[string]any, 0, len(hits))
	for _, h := range hits {
		out = append(out, map[string]any{
			"y0": h.y0, "y1": h.y1, "x0": h.x0, "x1": h.x1,
			"focus": h.focus, "runStage": h.runStage,
		})
	}
	return out
}

func hitmapSummaryHits(hits []Hit) []map[string]any {
	out := make([]map[string]any, 0, len(hits))
	for _, h := range hits {
		out = append(out, map[string]any{
			"y0": h.Y0, "y1": h.Y1, "x0": h.X0, "x1": h.X1,
			"focus": h.Focus, "runStage": h.RunStage,
		})
	}
	return out
}

// mouseDebugBanner writes a session start line (call once at dashboard start).
func mouseDebugBanner(extra map[string]any) {
	if !mouseDebugEnabled() {
		return
	}
	if extra == nil {
		extra = map[string]any{}
	}
	extra["log"] = mouseDebugLogPath
	extra["pid"] = os.Getpid()
	// Also a human-readable first line for grepping.
	mouseDebugMu.Lock()
	defer mouseDebugMu.Unlock()
	f, err := os.OpenFile(mouseDebugLogPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if err != nil {
		return
	}
	_, _ = fmt.Fprintf(f, "# wrk tui mouse debug start pid=%d %s\n", os.Getpid(), time.Now().Format(time.RFC3339))
	b, _ := json.Marshal(extra)
	if len(b) > 0 {
		_, _ = f.Write(append(b, '\n'))
	}
	_ = f.Close()
}
