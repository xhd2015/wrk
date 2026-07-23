package wrkcli

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRunCLIVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := RunCLI([]string{"--version"}, RunOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if code != 0 {
		t.Fatalf("exit %d stderr=%q", code, stderr.String())
	}
	if !strings.HasPrefix(stdout.String(), "v") || !strings.HasSuffix(stdout.String(), "\n") {
		t.Fatalf("stdout=%q", stdout.String())
	}
	if stderr.Len() != 0 {
		t.Fatalf("stderr should be empty, got %q", stderr.String())
	}
}

func TestRunCLIHelpMentionsVersion(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := RunCLI([]string{"-h"}, RunOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if code != 0 {
		t.Fatalf("exit %d stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), "--version") {
		t.Fatalf("help missing --version:\n%s", stdout.String())
	}
}

func TestRunCLISkillListWithWrkHomeOverride(t *testing.T) {
	home := t.TempDir()
	var stdout, stderr bytes.Buffer
	code := RunCLI([]string{"skill", "--list"}, RunOptions{
		Stdout:  &stdout,
		Stderr:  &stderr,
		WrkHome: home,
		Dir:     t.TempDir(),
	})
	if code != 0 {
		t.Fatalf("exit %d stderr=%q", code, stderr.String())
	}
	if strings.TrimSpace(stdout.String()) != "wrk" {
		t.Fatalf("stdout=%q", stdout.String())
	}
	// Override must not require process WRK_HOME.
	if os.Getenv("WRK_HOME") == home {
		t.Fatal("RunCLI must not Setenv WRK_HOME")
	}
}

func TestRunCLIVersionMutex(t *testing.T) {
	var stdout, stderr bytes.Buffer
	code := RunCLI([]string{"--version", "--list"}, RunOptions{
		Stdout: &stdout,
		Stderr: &stderr,
	})
	if code == 0 {
		t.Fatal("expected non-zero exit for --version --list")
	}
	if stdout.Len() != 0 {
		t.Fatalf("stdout should be empty, got %q", stdout.String())
	}
	lower := strings.ToLower(stderr.String())
	if !strings.Contains(lower, "mutually exclusive") && !strings.Contains(lower, "unexpected") {
		t.Fatalf("stderr should mention mutual exclusion, got %q", stderr.String())
	}
}

func TestRunCLISetConfigShowEmpty(t *testing.T) {
	home := filepath.Join(t.TempDir(), ".wrk")
	if err := os.MkdirAll(home, 0o755); err != nil {
		t.Fatal(err)
	}
	var stdout, stderr bytes.Buffer
	code := RunCLI([]string{"--set-config", "--show"}, RunOptions{
		Stdout:  &stdout,
		Stderr:  &stderr,
		WrkHome: home,
		Dir:     t.TempDir(),
	})
	if code != 0 {
		t.Fatalf("exit %d stderr=%q", code, stderr.String())
	}
	if strings.TrimSpace(stdout.String()) != "{}" {
		t.Fatalf("stdout=%q", stdout.String())
	}
}

func TestRunCLIRejectsProcessIsolationFields(t *testing.T) {
	var stderr bytes.Buffer
	code := RunCLI([]string{"--version"}, RunOptions{
		Stderr:   &stderr,
		ExtraEnv: []string{"FOO=bar"},
	})
	if code != 2 {
		t.Fatalf("exit %d want 2 stderr=%q", code, stderr.String())
	}
	if !strings.Contains(stderr.String(), "product-binary isolation") {
		t.Fatalf("stderr=%q", stderr.String())
	}
}
