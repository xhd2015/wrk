package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/xhd2015/wrk/wrkcli/unwind"
)

const jsonPlaceholder = "/*__UNWIND_JSON__*/null"

const (
	statusLoading = "loading"
	statusReady   = "ready"
	statusError   = "error"
)

// Options configures wrk --unwind --web.
type Options struct {
	WorkDir string
	WrkHome string
	Port    int
	Flags   unwind.UnwindFlags
	Host    unwind.Host
}

type pagePayload struct {
	Status string          `json:"status"`
	Error  string          `json:"error,omitempty"`
	Plan   *unwind.JobPlan `json:"plan"`
}

// planOut is /plan JSON: status + flattened JobPlan fields.
type planOut struct {
	Status string `json:"status"`
	Error  string `json:"error,omitempty"`
	*unwind.JobPlan
}

type server struct {
	opts unwind.UnwindFlags
	host unwind.Host
	dir  string
	home string

	mu       sync.Mutex
	status   string // loading | ready | error
	snapErr  string
	snap     *unwind.Snapshot
	running  bool
	logs     []string
	subs     []chan string
	doneCh   chan string
}

// Serve listens immediately, prints the URL, then collects the snapshot in the
// background so the UI can show a loading state on large stacks.
func Serve(opts Options) error {
	s := newServer(opts)
	s.startSnapshotAsync()
	h := s.mux()

	ln, port, err := listenLocal(opts.Port)
	if err != nil {
		return err
	}
	fmt.Printf("http://127.0.0.1:%d/\n", port)

	httpSrv := &http.Server{Handler: h, ReadHeaderTimeout: 10 * time.Second}
	errCh := make(chan error, 1)
	go func() {
		err := httpSrv.Serve(ln)
		if err == nil || err == http.ErrServerClosed {
			errCh <- nil
			return
		}
		errCh <- err
	}()

	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(sigCh)

	select {
	case err := <-errCh:
		return err
	case <-sigCh:
		_ = httpSrv.Close()
		<-errCh
		return nil
	}
}

// NewHandler builds the unwind preview mux with a synchronous snapshot (tests).
func NewHandler(opts Options) (http.Handler, error) {
	s := newServer(opts)
	if err := s.collectSnapshot(); err != nil {
		return nil, err
	}
	return s.mux(), nil
}

func newServer(opts Options) *server {
	return &server{
		opts:   opts.Flags,
		host:   opts.Host,
		dir:    opts.WorkDir,
		home:   opts.WrkHome,
		status: statusLoading,
	}
}

func (s *server) mux() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/", s.handleIndex)
	mux.HandleFunc("/plan", s.handlePlan)
	mux.HandleFunc("/run", s.handleRun)
	mux.HandleFunc("/run/logs", s.handleLogs)
	return mux
}

func (s *server) startSnapshotAsync() {
	go func() {
		_ = s.collectSnapshot()
	}()
}

func (s *server) collectSnapshot() error {
	restore := unwind.UseHost(s.host)
	defer restore()
	snap, err := unwind.CollectSnapshot(s.dir, unwind.SnapshotOpts{Cascade: true})
	s.mu.Lock()
	defer s.mu.Unlock()
	if err != nil {
		s.status = statusError
		s.snapErr = err.Error()
		s.snap = nil
		return err
	}
	s.snap = snap
	s.snapErr = ""
	s.status = statusReady
	return nil
}

func (s *server) snapState() (status, errMsg string, snap *unwind.Snapshot) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.status, s.snapErr, s.snap
}

func (s *server) currentPlan(flags unwind.UnwindFlags) planOut {
	status, errMsg, snap := s.snapState()
	restore := unwind.UseHost(s.host)
	defer restore()
	plan := unwind.BuildJobPlan(snap, flags)
	if status == statusLoading {
		plan.Blockers = nil
		plan.CanRun = false
		plan.Epochs = nil
		plan.Graph = nil
		plan.ApplyHint = ""
	}
	if status == statusError {
		plan.CanRun = false
		if errMsg != "" {
			plan.Blockers = []string{errMsg}
		}
		plan.Epochs = nil
		plan.Graph = nil
	}
	return planOut{Status: status, Error: errMsg, JobPlan: plan}
}

func (s *server) handleIndex(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	out := s.currentPlan(s.opts)
	raw, _ := json.Marshal(pagePayload{Status: out.Status, Error: out.Error, Plan: out.JobPlan})
	body := strings.Replace(indexHTML, jsonPlaceholder, "/*__UNWIND_JSON__*/"+string(raw), 1)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = io.WriteString(w, body)
}

func (s *server) handlePlan(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet, http.MethodHead:
		flags := flagsFromQuery(r, unwind.UnwindFlags{})
		writeJSON(w, http.StatusOK, s.currentPlan(flags))
	case http.MethodPost:
		var jf unwind.JobFlags
		if r.Body != nil {
			dec := json.NewDecoder(r.Body)
			if err := dec.Decode(&jf); err != nil && err != io.EOF {
				writeJSON(w, http.StatusBadRequest, map[string]string{"error": "invalid job flags JSON"})
				return
			}
		}
		flags := unwind.FlagsFromJob(jf, s.opts.GenCommitArgs)
		writeJSON(w, http.StatusOK, s.currentPlan(flags))
	default:
		w.Header().Set("Allow", "GET, HEAD, POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

func (s *server) handleRun(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		w.Header().Set("Allow", "POST")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	status, errMsg, snap := s.snapState()
	if status == statusLoading {
		writeJSON(w, http.StatusServiceUnavailable, map[string]string{"error": "plan still loading"})
		return
	}
	if status == statusError || snap == nil {
		msg := "snapshot unavailable"
		if errMsg != "" {
			msg = errMsg
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": msg})
		return
	}
	var jf unwind.JobFlags
	if r.Body != nil {
		_ = json.NewDecoder(r.Body).Decode(&jf)
	}
	flags := unwind.FlagsFromJob(jf, s.opts.GenCommitArgs)
	plan := s.currentPlan(flags)
	if !plan.CanRun {
		msg := "cannot run"
		if plan.JobPlan != nil && len(plan.Blockers) > 0 {
			msg = plan.Blockers[0]
		}
		writeJSON(w, http.StatusBadRequest, map[string]string{"error": msg})
		return
	}

	s.mu.Lock()
	if s.running {
		s.mu.Unlock()
		writeJSON(w, http.StatusConflict, map[string]string{"error": "unwind already running"})
		return
	}
	s.running = true
	s.logs = nil
	s.doneCh = make(chan string, 1)
	s.mu.Unlock()

	go s.apply(flags, snap)
	w.WriteHeader(http.StatusAccepted)
}

var applyStdioMu sync.Mutex

func (s *server) apply(flags unwind.UnwindFlags, snap *unwind.Snapshot) {
	restore := unwind.UseHost(s.host)
	defer restore()
	s.appendLog("==== unwind apply ====")
	out, err := captureApplyOutput(func() error {
		return unwind.ApplyUnwind(s.dir, s.home, snap.Inv.Members, snap.RepoEdges, snap.Peel, flags)
	})
	for _, line := range strings.Split(out, "\n") {
		if line != "" {
			s.appendLog(line)
		}
	}
	done := "ok"
	if err != nil {
		s.appendLog("Error: " + err.Error())
		done = "error: " + err.Error()
	} else {
		s.appendLog("unwind: done")
		if next, serr := unwind.CollectSnapshot(s.dir, unwind.SnapshotOpts{Cascade: true}); serr == nil {
			s.mu.Lock()
			s.snap = next
			s.status = statusReady
			s.snapErr = ""
			s.mu.Unlock()
		}
	}
	s.mu.Lock()
	s.running = false
	for _, ch := range s.subs {
		select {
		case ch <- "DONE " + done:
		default:
		}
	}
	s.mu.Unlock()
	select {
	case s.doneCh <- done:
	default:
	}
}

func (s *server) handleLogs(w http.ResponseWriter, r *http.Request) {
	fl, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming unsupported", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	ch := make(chan string, 32)
	s.mu.Lock()
	for _, line := range s.logs {
		fmt.Fprintf(w, "event: log\ndata: %s\n\n", line)
	}
	if !s.running {
		s.mu.Unlock()
		fmt.Fprintf(w, "event: done\ndata: ok\n\n")
		fl.Flush()
		return
	}
	s.subs = append(s.subs, ch)
	s.mu.Unlock()
	fl.Flush()

	notify := r.Context().Done()
	for {
		select {
		case <-notify:
			return
		case msg := <-ch:
			if strings.HasPrefix(msg, "DONE ") {
				fmt.Fprintf(w, "event: done\ndata: %s\n\n", strings.TrimPrefix(msg, "DONE "))
				fl.Flush()
				return
			}
			fmt.Fprintf(w, "event: log\ndata: %s\n\n", msg)
			fl.Flush()
		}
	}
}

func (s *server) appendLog(line string) {
	s.mu.Lock()
	s.logs = append(s.logs, line)
	for _, ch := range s.subs {
		select {
		case ch <- line:
		default:
		}
	}
	s.mu.Unlock()
}

func flagsFromQuery(r *http.Request, base unwind.UnwindFlags) unwind.UnwindFlags {
	q := r.URL.Query()
	boolQ := func(key string, cur bool) bool {
		if !q.Has(key) {
			return cur
		}
		v := q.Get(key)
		return v == "1" || v == "true" || v == "on"
	}
	// Seed from full UnwindFlags (peels --commit/--add-all out of GenCommitArgs into JobFlags).
	jf := unwind.BuildJobPlan(nil, base).Flags
	jf.MergeBack = boolQ("merge_back", jf.MergeBack)
	jf.Done = boolQ("done", jf.Done)
	jf.TagNext = boolQ("tag_next", jf.TagNext)
	jf.Push = boolQ("push", jf.Push)
	jf.Sync = boolQ("sync", jf.Sync)
	jf.ReinstallLocal = boolQ("reinstall_local", jf.ReinstallLocal)
	jf.AddAll = boolQ("add_all", jf.AddAll)
	jf.GenCommitMsg = boolQ("gen_commit_msg", jf.GenCommitMsg)
	jf.Commit = boolQ("commit", jf.Commit)
	jf.NoVerify = boolQ("no_verify", jf.NoVerify)
	return unwind.FlagsFromJob(jf, base.GenCommitArgs)
}

func captureApplyOutput(fn func() error) (string, error) {
	applyStdioMu.Lock()
	defer applyStdioMu.Unlock()
	pr, pw, err := os.Pipe()
	if err != nil {
		return "", fn()
	}
	oldOut, oldErr := os.Stdout, os.Stderr
	os.Stdout, os.Stderr = pw, pw
	var buf bytes.Buffer
	done := make(chan struct{})
	go func() {
		_, _ = io.Copy(&buf, pr)
		close(done)
	}()
	runErr := fn()
	_ = pw.Close()
	os.Stdout, os.Stderr = oldOut, oldErr
	<-done
	_ = pr.Close()
	return buf.String(), runErr
}

func writeJSON(w http.ResponseWriter, code int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(v)
}

func listenLocal(port int) (net.Listener, int, error) {
	if port < 0 {
		return nil, 0, fmt.Errorf("wrk: invalid --port %d", port)
	}
	if port > 0 {
		ln, err := net.Listen("tcp", ":"+strconv.Itoa(port))
		if err != nil {
			return nil, 0, fmt.Errorf("wrk: listen :%d: %w", port, err)
		}
		return ln, port, nil
	}
	for p := 8080; p < 8080+200; p++ {
		ln, err := net.Listen("tcp", ":"+strconv.Itoa(p))
		if err == nil {
			return ln, p, nil
		}
	}
	ln, err := net.Listen("tcp", ":0")
	if err != nil {
		return nil, 0, fmt.Errorf("wrk: listen :0: %w", err)
	}
	addr, ok := ln.Addr().(*net.TCPAddr)
	if !ok {
		_ = ln.Close()
		return nil, 0, fmt.Errorf("wrk: listen: unexpected address type %T", ln.Addr())
	}
	return ln, addr.Port, nil
}
