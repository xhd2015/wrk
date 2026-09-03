package wrkcli

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func TestNormalizeCommitMessage(t *testing.T) {
	t.Parallel()
	cases := []struct {
		in, want string
	}{
		{"feat: x", "feat: x"},
		{"feat: x\n", "feat: x"},
		{"feat: x\n\n", "feat: x"},
		{"feat: x\n\nbody\n", "feat: x\n\nbody"},
		{"  trailing spaces  \n", "  trailing spaces"},
	}
	for _, tc := range cases {
		if got := normalizeCommitMessage(tc.in); got != tc.want {
			t.Fatalf("normalizeCommitMessage(%q)=%q want %q", tc.in, got, tc.want)
		}
	}
}

func TestIsNoStagedCommitErr(t *testing.T) {
	t.Parallel()
	if isNoStagedCommitErr(nil) {
		t.Fatal("nil should not match")
	}
	if !isNoStagedCommitErr(errors.New("no staged changes to commit")) {
		t.Fatal("manual empty-index error should match")
	}
	if !isNoStagedCommitErr(errors.New("no staged changes to generate commit message for")) {
		t.Fatal("gen-commit empty-index error should match")
	}
	if isNoStagedCommitErr(errors.New("git commit failed: hook")) {
		t.Fatal("unrelated error must not match")
	}
}

func TestManualCommitMessageMatchesHEAD(t *testing.T) {
	t.Parallel()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
	dir := t.TempDir()
	run := func(args ...string) {
		t.Helper()
		cmd := exec.Command("git", args...)
		cmd.Dir = dir
		cmd.Env = append(os.Environ(),
			"GIT_CONFIG_NOSYSTEM=1",
			"GIT_CONFIG_GLOBAL=/dev/null",
		)
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %s: %v\n%s", strings.Join(args, " "), err, out)
		}
	}
	run("init", "--template=", "-b", "main")
	run("config", "user.email", "test@example.com")
	run("config", "user.name", "Test")
	run("config", "core.hooksPath", "/dev/null")
	if err := os.WriteFile(filepath.Join(dir, "f.txt"), []byte("x\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	run("add", "f.txt")
	run("commit", "-m", "feat: subject\n\nbody line\n")

	if !manualCommitMessageMatchesHEAD(dir, "feat: subject\n\nbody line") {
		t.Fatal("expected match ignoring trailing newline from %B")
	}
	if !manualCommitMessageMatchesHEAD(dir, "feat: subject\n\nbody line\n") {
		t.Fatal("expected match with trailing newline on -m")
	}
	if manualCommitMessageMatchesHEAD(dir, "feat: other") {
		t.Fatal("different message must not match")
	}
	if manualCommitMessageMatchesHEAD(filepath.Join(dir, "missing"), "feat: subject") {
		t.Fatal("bad workDir must not match")
	}
}
