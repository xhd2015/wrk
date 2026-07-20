package wrkcli

import (
	"bytes"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestFormatFileCountPreview(t *testing.T) {
	cases := []struct {
		n    int
		want string
	}{
		{0, "clean"},
		{-1, "clean"},
		{1, "1 file"},
		{3, "3 files"},
	}
	for _, tc := range cases {
		if got := formatFileCountPreview(tc.n); got != tc.want {
			t.Errorf("formatFileCountPreview(%d)=%q want %q", tc.n, got, tc.want)
		}
	}
}

func TestFormatAheadPreview(t *testing.T) {
	if got := formatAheadPreview(0); got != "ahead 0" {
		t.Errorf("got %q", got)
	}
	if got := formatAheadPreview(7); got != "ahead 7" {
		t.Errorf("got %q", got)
	}
}

func TestFormatPushPreview(t *testing.T) {
	cases := []struct {
		ahead, behind int
		want          string
	}{
		{0, 0, "up to date"},
		{2, 0, "ahead 2"},
		{0, 3, "behind 3"},
		{1, 2, "ahead 1, behind 2"},
	}
	for _, tc := range cases {
		if got := formatPushPreview(tc.ahead, tc.behind); got != tc.want {
			t.Errorf("formatPushPreview(%d,%d)=%q want %q", tc.ahead, tc.behind, got, tc.want)
		}
	}
}

func TestCountNonEmptyLines(t *testing.T) {
	in := " M a.go\n?? b.go\n\n"
	if n := countNonEmptyLines(in); n != 2 {
		t.Fatalf("count=%d want 2", n)
	}
	if n := countNonEmptyLines(""); n != 0 {
		t.Fatalf("empty count=%d", n)
	}
}

func TestDashboardStagePreviewEmptyOnUnknown(t *testing.T) {
	preview, logs := dashboardStagePreview("/tmp", "not-a-stage")
	if preview != "" || len(logs) != 0 {
		t.Fatalf("unknown stage → empty, got %q logs=%v", preview, logs)
	}
	preview, logs = dashboardStagePreview("", "add-changes")
	if preview != "" || len(logs) != 0 {
		t.Fatalf("empty workDir → empty, got %q logs=%v", preview, logs)
	}
}

// TestPreviewPushUpstreamNoUpstreamToLogs ensures missing @{u} does not leak
// git fatals onto process stderr; diagnostics are returned as normal log lines.
func TestPreviewPushUpstreamNoUpstreamToLogs(t *testing.T) {
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_CONFIG_NOSYSTEM=1",
			"GIT_CONFIG_GLOBAL=/dev/null",
			"GIT_AUTHOR_NAME=test",
			"GIT_AUTHOR_EMAIL=test@example.com",
			"GIT_COMMITTER_NAME=test",
			"GIT_COMMITTER_EMAIL=test@example.com",
		)
		out, err := cmd.CombinedOutput()
		if err != nil {
			t.Fatalf("git %v: %v\n%s", args, err, out)
		}
	}
	run("init", "-b", "master-no-upstream-test")
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "f.txt")
	run("commit", "-m", "init")

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldErr := os.Stderr
	os.Stderr = w
	preview, logs := previewPushUpstream(dir)
	_ = w.Close()
	os.Stderr = oldErr
	leaked, _ := io.ReadAll(r)
	_ = r.Close()

	if preview != "" {
		t.Fatalf("no upstream → empty preview, got %q", preview)
	}
	if bytes.Contains(leaked, []byte("fatal")) ||
		bytes.Contains(leaked, []byte("upstream")) {
		t.Fatalf("git fatal leaked to process stderr:\n%s", leaked)
	}
	if s := strings.TrimSpace(string(leaked)); s != "" {
		t.Fatalf("unexpected process stderr:\n%q", s)
	}
	// Error content must be preserved as normal logs for the TUI Log panel.
	joined := strings.Join(logs, "\n")
	if !strings.Contains(joined, "upstream") && !strings.Contains(joined, "fatal") {
		t.Fatalf("expected upstream/fatal in logs for Log panel, got %v", logs)
	}
}

func TestGitOutputDirCaptureNoTTYLeak(t *testing.T) {
	dir := t.TempDir()
	cmd := exec.Command("git", "init", "-b", "main")
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("init: %v\n%s", err, out)
	}
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	oldErr := os.Stderr
	os.Stderr = w
	_, stderr, runErr := gitOutputDirCapture(dir, "rev-parse", "refs/does-not-exist")
	_ = w.Close()
	os.Stderr = oldErr
	leaked, _ := io.ReadAll(r)
	_ = r.Close()
	if runErr == nil {
		t.Fatal("expected error for missing ref")
	}
	if bytes.Contains(leaked, []byte("fatal")) || bytes.Contains(leaked, []byte("does-not-exist")) {
		t.Fatalf("stderr leaked to process:\n%s", leaked)
	}
	if !strings.Contains(stderr, "fatal") && !strings.Contains(stderr, "does-not-exist") {
		t.Fatalf("expected captured stderr content, got %q", stderr)
	}
}

func TestSplitCapturedLogLines(t *testing.T) {
	got := splitCapturedLogLines("fatal: no upstream\n\nwarning: x\n")
	if len(got) != 2 || got[0] != "fatal: no upstream" || got[1] != "warning: x" {
		t.Fatalf("got %v", got)
	}
}
