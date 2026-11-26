package workops

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// hookRepo builds a repository whose generated stamp file is rewritten on both
// sides, plus a linked worktree on the feature branch: rebasing that worktree
// onto main can only conflict, which is the case the on-conflict hook exists
// for. withHook decides whether the repository configures a resolver.
func hookRepo(t *testing.T, withHook bool) (main, worktree string) {
	t.Helper()
	main = t.TempDir()
	runReclaimGit(t, main, "init", "-b", "main")
	runReclaimGit(t, main, "config", "user.email", "t@t")
	runReclaimGit(t, main, "config", "user.name", "t")
	if withHook {
		// The resolver lives in history, the way a real repository's hook does.
		writeHookFile(t, main, "script/git/hooks/rebase/on-conflict.sh", `#!/usr/bin/env bash
set -euo pipefail
root=$(git rev-parse --show-toplevel)
printf '%s+1' "$(git rev-parse HEAD)" > "$root/REVISION.txt"
git -C "$root" add -- REVISION.txt
`)
	}
	writeHookFile(t, main, "REVISION.txt", "base")
	runReclaimGit(t, main, "add", ".")
	runReclaimGit(t, main, "commit", "-m", "init")

	worktree = filepath.Join(t.TempDir(), "topic-hook")
	runReclaimGit(t, main, "worktree", "add", "-b", "topic-hook", worktree)
	writeHookFile(t, worktree, "REVISION.txt", "feature")
	writeHookFile(t, worktree, "feature.txt", "feature\n")
	runReclaimGit(t, worktree, "add", ".")
	runReclaimGit(t, worktree, "commit", "-m", "feature")

	writeHookFile(t, main, "REVISION.txt", "main")
	writeHookFile(t, main, "main.txt", "main\n")
	runReclaimGit(t, main, "add", ".")
	runReclaimGit(t, main, "commit", "-m", "main")
	return main, worktree
}

func writeHookFile(t *testing.T, dir, rel, content string) {
	t.Helper()
	path := filepath.Join(dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	mode := os.FileMode(0o644)
	if strings.HasSuffix(rel, ".sh") {
		mode = 0o755
	}
	if err := os.WriteFile(path, []byte(content), mode); err != nil {
		t.Fatal(err)
	}
}

func hookRead(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestMergeBackResolvesConflictWithRepositoryHook(t *testing.T) {
	main, worktree := hookRepo(t, true)
	result, err := MergeBackFull(context.Background(), MergeBackOptions{
		WorktreeDir: worktree,
		WrkHome:     t.TempDir(),
	})
	if err != nil {
		t.Fatalf("merge-back: %v", err)
	}
	if result.Action != "rebased-and-merged" {
		t.Fatalf("Action=%q message=%q, want rebased-and-merged", result.Action, result.Message)
	}
	if !strings.Contains(result.Message, "on-conflict hook") {
		t.Fatalf("Message=%q, want it to name the resolver", result.Message)
	}
	// Both sides landed, and the stamp is the value the hook generated.
	for _, name := range []string{"feature.txt", "main.txt"} {
		if _, statErr := os.Stat(filepath.Join(main, name)); statErr != nil {
			t.Fatalf("%s missing from main: %v", name, statErr)
		}
	}
	wantStamp := runReclaimGitOutput(t, main, "rev-parse", "HEAD^") + "+1"
	if got := hookRead(t, filepath.Join(main, "REVISION.txt")); got != wantStamp {
		t.Fatalf("REVISION.txt = %q, want %q", got, wantStamp)
	}
	if dirty := runReclaimGitOutput(t, main, "status", "--porcelain"); dirty != "" {
		t.Fatalf("main is dirty after land: %q", dirty)
	}
}

func TestMergeBackWithoutHookStillReportsConflict(t *testing.T) {
	_, worktree := hookRepo(t, false)
	before := runReclaimGitOutput(t, worktree, "rev-parse", "HEAD")

	_, err := MergeBackFull(context.Background(), MergeBackOptions{
		WorktreeDir: worktree,
		WrkHome:     t.TempDir(),
	})
	if err == nil {
		t.Fatal("a conflicted rebase without a resolver must fail the land")
	}
	if !strings.Contains(err.Error(), "rebase conflict") {
		t.Fatalf("err=%v, want the rebase-conflict contract", err)
	}
	// The previous contract is preserved: the rebase is aborted, not left half
	// applied, so the operator can inspect the worktree as it was.
	if after := runReclaimGitOutput(t, worktree, "rev-parse", "HEAD"); after != before {
		t.Fatalf("worktree HEAD moved to %s, want %s", after, before)
	}
	if got := runReclaimGitOutput(t, worktree, "status", "--porcelain"); got != "" {
		t.Fatalf("worktree left dirty after the aborted rebase: %q", got)
	}
}

// runReclaimGitOutput is runReclaimGit for a command whose output the test needs.
func runReclaimGitOutput(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(), "GIT_OPTIONAL_LOCKS=0", "GIT_TERMINAL_PROMPT=0")
	out, err := cmd.Output()
	if err != nil {
		t.Fatalf("git %s: %v", strings.Join(args, " "), err)
	}
	return strings.TrimSpace(string(out))
}
