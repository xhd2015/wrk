package wrkserver

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func TestCreateOpAndStreamLogs(t *testing.T) {
	s := New(Options{})
	mux := http.NewServeMux()
	s.Register(mux, "/api/wrk")

	body := `{"action":"sync","target_id":"wt-1"}`
	req := httptest.NewRequest(http.MethodPost, "/api/wrk/ops", strings.NewReader(body))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusAccepted {
		t.Fatalf("POST /ops status=%d body=%s", rec.Code, rec.Body.String())
	}
	var created CreateOpResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &created); err != nil {
		t.Fatal(err)
	}
	if created.OpID == "" || created.Action != "sync" {
		t.Fatalf("unexpected create response: %+v", created)
	}

	// Stream until done (mock finishes in ~2s).
	logReq := httptest.NewRequest(http.MethodGet, "/api/wrk/ops/"+created.OpID+"/logs", nil)
	logRec := httptest.NewRecorder()

	done := make(chan struct{})
	go func() {
		mux.ServeHTTP(logRec, logReq)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(8 * time.Second):
		t.Fatal("SSE stream timed out")
	}

	raw := logRec.Body.String()
	if !strings.Contains(raw, "event: log") {
		t.Fatalf("expected log events, got: %s", raw)
	}
	if !strings.Contains(raw, "event: done") {
		t.Fatalf("expected done event, got: %s", raw)
	}
	if !strings.Contains(raw, "starting sync") {
		t.Fatalf("expected starting message, got: %s", raw)
	}

	// Parse at least one log data line.
	sc := bufio.NewScanner(bytes.NewReader(logRec.Body.Bytes()))
	var sawLogData bool
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "data: ") && strings.Contains(line, "message") {
			sawLogData = true
			break
		}
	}
	if !sawLogData {
		t.Fatalf("no log data payloads in stream: %s", raw)
	}
}

func TestCreateOpRequiresAction(t *testing.T) {
	s := New(Options{})
	mux := http.NewServeMux()
	s.Register(mux, "/api/wrk")
	req := httptest.NewRequest(http.MethodPost, "/api/wrk/ops", strings.NewReader(`{}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d", rec.Code)
	}
}

func TestCreateOpUnknownAction(t *testing.T) {
	s := New(Options{})
	mux := http.NewServeMux()
	s.Register(mux, "/api/wrk")
	req := httptest.NewRequest(http.MethodPost, "/api/wrk/ops", strings.NewReader(`{"action":"explode"}`))
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("status=%d body=%s", rec.Code, rec.Body.String())
	}
	_ = io.Discard
}
