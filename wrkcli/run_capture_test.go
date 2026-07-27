package wrkcli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestCaptureVersion(t *testing.T) {
	res := RunCapture("--version")
	if res.ExitCode != 0 {
		t.Fatalf("exit %d stderr=%q", res.ExitCode, res.Stderr)
	}
	want := Version() + "\n"
	if res.Stdout != want {
		t.Fatalf("stdout=%q want %q", res.Stdout, want)
	}
	if res.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", res.Stderr)
	}
}

func TestCaptureSkillList(t *testing.T) {
	home := t.TempDir()
	res := Capture(CaptureOpts{
		Args: []string{"skill", "--list"},
		Env:  []string{"WRK_HOME=" + home},
	})
	if res.ExitCode != 0 {
		t.Fatalf("exit %d stderr=%q stdout=%q", res.ExitCode, res.Stderr, res.Stdout)
	}
	if res.Stdout != "wrk\n" {
		t.Fatalf("stdout=%q want wrk\\n", res.Stdout)
	}
}

func TestCaptureVersionMutexWithOtherFlags(t *testing.T) {
	res := RunCapture("--version", "--list")
	if res.ExitCode == 0 {
		t.Fatal("expected non-zero exit for --version with --list")
	}
	if res.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", res.Stdout)
	}
	if !strings.Contains(strings.ToLower(res.Stderr), "mutually exclusive") &&
		!strings.Contains(strings.ToLower(res.Stderr), "unexpected") {
		t.Fatalf("stderr should mention exclusion, got %q", res.Stderr)
	}
}

func TestCaptureSkillInstallDryRunUsesDir(t *testing.T) {
	work := t.TempDir()
	home := filepath.Join(work, ".wrk")
	cwd := filepath.Join(work, "workspace")
	if err := os.MkdirAll(cwd, 0o755); err != nil {
		t.Fatal(err)
	}
	res := Capture(CaptureOpts{
		Args: []string{"skill", "--install", "--cursor", "--dry-run"},
		Dir:  cwd,
		Env:  []string{"WRK_HOME=" + home},
	})
	if res.ExitCode != 0 {
		t.Fatalf("exit %d stderr=%q stdout=%q", res.ExitCode, res.Stderr, res.Stdout)
	}
	if !strings.Contains(res.Stdout, ".cursor/skills/wrk") {
		t.Fatalf("stdout missing install path, got %q", res.Stdout)
	}
	if _, err := os.Stat(filepath.Join(cwd, ".cursor", "skills", "wrk")); !os.IsNotExist(err) {
		t.Fatalf("dry-run should not create skill dir, err=%v", err)
	}
}

func TestCaptureStdinIsPipeWhenSet(t *testing.T) {
	// Empty Stdin must not swap os.Stdin; non-empty must yield a readable pipe.
	old := os.Stdin
	t.Cleanup(func() { os.Stdin = old })

	// Direct unit of applyStdin (Capture holds the mutex; exercise helper path).
	restore, err := applyStdin("n\n")
	if err != nil {
		t.Fatal(err)
	}
	defer restore()
	if os.Stdin == old {
		t.Fatal("expected os.Stdin replaced with pipe")
	}
	buf := make([]byte, 8)
	n, err := os.Stdin.Read(buf)
	if err != nil {
		t.Fatalf("read stdin: %v", err)
	}
	if string(buf[:n]) != "n\n" {
		t.Fatalf("stdin data=%q want n\\n", string(buf[:n]))
	}
}
