package wrkserver

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"time"
)

// CreateOpRequest is the POST {base}/ops body.
type CreateOpRequest struct {
	Action   string `json:"action"`
	TargetID string `json:"target_id,omitempty"`
	Label    string `json:"label,omitempty"`
}

// CreateOpResponse is returned when a mock op is started.
type CreateOpResponse struct {
	OpID   string `json:"op_id"`
	Action string `json:"action"`
}

// LogLine is one streamed log event payload.
type LogLine struct {
	TS      string `json:"ts"`
	Level   string `json:"level"`
	Message string `json:"message"`
}

// OpDonePayload is the final SSE event.
type OpDonePayload struct {
	OK      bool   `json:"ok"`
	Summary string `json:"summary,omitempty"`
	Error   string `json:"error,omitempty"`
}

type mockOp struct {
	id       string
	action   string
	targetID string
	mu       sync.Mutex
	lines    []LogLine
	done     bool
	ok       bool
	summary  string
	errMsg   string
	subs     []chan logEvent
}

type logEvent struct {
	kind string // "log" | "done"
	line LogLine
	done OpDonePayload
}

var (
	opsMu   sync.Mutex
	ops sequencableOps
)

type sequencableOps struct {
	seq int
	m   map[string]*mockOp
}

func init() {
	ops.m = make(map[string]*mockOp)
}

// CreateOp handles POST {base}/ops — starts a mock streaming operation.
func (s *Server) CreateOp(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	var req CreateOpRequest
	if r.Body != nil {
		body, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "failed to read body")
			return
		}
		if len(strings.TrimSpace(string(body))) > 0 {
			if err := json.Unmarshal(body, &req); err != nil {
				writeError(w, http.StatusBadRequest, "invalid JSON: "+err.Error())
				return
			}
		}
	}
	action := strings.TrimSpace(req.Action)
	if action == "" {
		writeError(w, http.StatusBadRequest, "action is required")
		return
	}
	if !validMockAction(action) {
		writeError(w, http.StatusBadRequest, "unknown action: "+action)
		return
	}

	op := newMockOp(action, strings.TrimSpace(req.TargetID), strings.TrimSpace(req.Label))
	opsMu.Lock()
	ops.seq++
	op.id = fmt.Sprintf("op-%d", ops.seq)
	ops.m[op.id] = op
	opsMu.Unlock()

	go op.runMock()

	writeJSON(w, http.StatusAccepted, CreateOpResponse{OpID: op.id, Action: action})
}

// StreamOpLogs handles GET {base}/ops/{id}/logs as Server-Sent Events.
func (s *Server) StreamOpLogs(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		writeError(w, http.StatusMethodNotAllowed, "method not allowed")
		return
	}
	id := strings.TrimSpace(r.PathValue("id"))
	if id == "" {
		// Fallback for routers without PathValue.
		id = extractOpIDFromLogsPath(r.URL.Path)
	}
	if id == "" {
		writeError(w, http.StatusBadRequest, "op id required")
		return
	}

	opsMu.Lock()
	op := ops.m[id]
	opsMu.Unlock()
	if op == nil {
		writeError(w, http.StatusNotFound, "op not found")
		return
	}

	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, "streaming unsupported")
		return
	}

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	if r.Method == http.MethodHead {
		return
	}
	flusher.Flush()

	ch := op.subscribe()
	defer op.unsubscribe(ch)

	// Replay buffered lines so late subscribers catch up.
	for _, line := range op.snapshotLines() {
		if err := writeSSE(w, flusher, "log", line); err != nil {
			return
		}
	}
	if done, finished := op.snapshotDone(); finished {
		_ = writeSSE(w, flusher, "done", done)
		return
	}

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-ch:
			if !ok {
				return
			}
			if ev.kind == "done" {
				_ = writeSSE(w, flusher, "done", ev.done)
				return
			}
			if err := writeSSE(w, flusher, "log", ev.line); err != nil {
				return
			}
		}
	}
}

func extractOpIDFromLogsPath(path string) string {
	// …/ops/{id}/logs
	parts := strings.Split(strings.Trim(path, "/"), "/")
	for i := 0; i+2 < len(parts); i++ {
		if parts[i] == "ops" && parts[i+2] == "logs" {
			return parts[i+1]
		}
	}
	return ""
}

func writeSSE(w http.ResponseWriter, flusher http.Flusher, event string, v any) error {
	data, err := json.Marshal(v)
	if err != nil {
		return err
	}
	if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event, data); err != nil {
		return err
	}
	flusher.Flush()
	return nil
}

func validMockAction(action string) bool {
	switch action {
	case "rebase", "sync", "tag", "push", "agent-run":
		return true
	default:
		return false
	}
}

func newMockOp(action, targetID, label string) *mockOp {
	return &mockOp{
		action:   action,
		targetID: targetID,
		lines:    nil,
		subs:     nil,
	}
}

func (op *mockOp) subscribe() chan logEvent {
	ch := make(chan logEvent, 32)
	op.mu.Lock()
	defer op.mu.Unlock()
	if op.done {
		// Caller will use snapshot; still return closed channel after send done.
		go func() {
			ch <- logEvent{kind: "done", done: OpDonePayload{OK: op.ok, Summary: op.summary, Error: op.errMsg}}
			close(ch)
		}()
		return ch
	}
	op.subs = append(op.subs, ch)
	return ch
}

func (op *mockOp) unsubscribe(ch chan logEvent) {
	op.mu.Lock()
	defer op.mu.Unlock()
	out := op.subs[:0]
	for _, s := range op.subs {
		if s != ch {
			out = append(out, s)
		}
	}
	op.subs = out
}

func (op *mockOp) snapshotLines() []LogLine {
	op.mu.Lock()
	defer op.mu.Unlock()
	out := make([]LogLine, len(op.lines))
	copy(out, op.lines)
	return out
}

func (op *mockOp) snapshotDone() (OpDonePayload, bool) {
	op.mu.Lock()
	defer op.mu.Unlock()
	if !op.done {
		return OpDonePayload{}, false
	}
	return OpDonePayload{OK: op.ok, Summary: op.summary, Error: op.errMsg}, true
}

func (op *mockOp) emitLog(level, msg string) {
	line := LogLine{
		TS:      time.Now().UTC().Format(time.RFC3339Nano),
		Level:   level,
		Message: msg,
	}
	op.mu.Lock()
	op.lines = append(op.lines, line)
	subs := append([]chan logEvent(nil), op.subs...)
	op.mu.Unlock()
	for _, ch := range subs {
		select {
		case ch <- logEvent{kind: "log", line: line}:
		default:
		}
	}
}

func (op *mockOp) finish(ok bool, summary, errMsg string) {
	done := OpDonePayload{OK: ok, Summary: summary, Error: errMsg}
	op.mu.Lock()
	op.done = true
	op.ok = ok
	op.summary = summary
	op.errMsg = errMsg
	subs := op.subs
	op.subs = nil
	op.mu.Unlock()
	for _, ch := range subs {
		select {
		case ch <- logEvent{kind: "done", done: done}:
		default:
		}
		close(ch)
	}
}

func (op *mockOp) runMock() {
	target := op.targetID
	if target == "" {
		target = "(default)"
	}
	label := op.action
	steps := mockSteps(op.action, target, label)
	op.emitLog("info", fmt.Sprintf("starting %s on %s (mock)", op.action, target))
	for _, step := range steps {
		time.Sleep(350 * time.Millisecond)
		op.emitLog(step.level, step.msg)
	}
	time.Sleep(200 * time.Millisecond)
	op.emitLog("info", fmt.Sprintf("%s finished (mock)", op.action))
	op.finish(true, fmt.Sprintf("%s completed", op.action), "")
}

type mockStep struct {
	level string
	msg   string
}

func mockSteps(action, target, label string) []mockStep {
	_ = label
	switch action {
	case "rebase":
		return []mockStep{
			{level: "info", msg: fmt.Sprintf("fetching base for worktree %s", target)},
			{level: "info", msg: "computing merge-base with Main"},
			{level: "info", msg: "git rebase Main (mock dry-run)"},
			{level: "info", msg: "rebase clean — no conflicts"},
		}
	case "sync":
		return []mockStep{
			{level: "info", msg: fmt.Sprintf("preparing merge-back from %s", target)},
			{level: "info", msg: "checking worktree clean"},
			{level: "info", msg: "merging branch into Main (mock)"},
			{level: "info", msg: "sync complete"},
		}
	case "tag":
		return []mockStep{
			{level: "info", msg: "reading HEAD on Main"},
			{level: "info", msg: "creating annotated tag (mock)"},
			{level: "info", msg: "tag ready locally"},
		}
	case "push":
		return []mockStep{
			{level: "info", msg: "resolving remote origin"},
			{level: "info", msg: "git push origin Main (mock)"},
			{level: "info", msg: "push accepted"},
		}
	case "agent-run":
		return []mockStep{
			{level: "info", msg: fmt.Sprintf("launching agent-run in %s", target)},
			{level: "info", msg: "session opened (mock)"},
			{level: "info", msg: "agent ready for prompts"},
		}
	default:
		return []mockStep{{level: "info", msg: "noop"}}
	}
}
