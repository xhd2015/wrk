# Scenario

**Feature**: wrk --status tolerates scan-discovered broken linked worktrees

```
# scan_repo discovers repos under checkout root in path order
wrk --status from main cwd -> scan_repo.Scan(root) -> status blocks per repo

# enrich failure on one scan row must not abort the run (same policy as appended broken)
scan-discovered repo (alive checkout, git fails) -> minimal relative Dir + Status: error: ...
```

## Preconditions

- Git must be available.
- Fixture layout: main repo `myrepo`, nested independent `tools/good`, nested main `vendor/host` with linked worktree `vendor/host/broken-wt`.
- Broken worktree: checkout dir exists; `.git` gitlink points at non-existent `gitdir` under `{WorkRoot}/stale-main/`.
- Scan-discovered broken blocks use **relative** `Dir` (not absolute like appended external worktrees).

## Steps

- `setupNestedBrokenLinkedFixture` commits root `.gitignore` with `tools/` and `vendor/` after
  root init so nested independent checkouts are not counted as untracked on the parent
  (root porcelain stays clean when untracked is included).
- Descendants call `setupNestedBrokenLinkedFixture` then run `wrk --status` from `{WorkRoot}/myrepo`.
- Color leaf adds `--color` to force red `error: …` on the broken block value.

## Context

- Healthy scan blocks unchanged (full `Dir`/`Branch`/`Commit`/`Status`; root block includes `Remote:` from main checkout cwd).
- Linked worktree blocks normally include `Master:` when healthy; broken blocks omit all fields except `Dir` and `Status`.
- Block order follows `scan_repo` path ordering.

```go
import (
	"fmt"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/gitops/git/git_isolated"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	req.Args = []string{"--status"}
	return nil
}

func setupNestedBrokenLinkedFixture(t *testing.T, req *Request) {
	t.Helper()
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	statusInitRepoWithSubject(t, mainRepo, "root repo")
	// Nested independent git dirs (tools/, vendor/) appear as ?? on the parent when
	// untracked files are included. Ignore them so the root block stays clean while
	// children are still discovered by scan_repo and report their own status.
	writeFile(t, filepath.Join(mainRepo, ".gitignore"), "tools/\nvendor/\n")
	runGitIsolated(t, mainRepo, "add", ".gitignore")
	runGitIsolated(t, mainRepo, "commit", "-m", "ignore nested tools and vendor")

	goodChild := filepath.Join(mainRepo, "tools", "good")
	statusInitRepoWithSubject(t, goodChild, "good child")

	hostMain := filepath.Join(mainRepo, "vendor", "host")
	statusInitRepoWithSubject(t, hostMain, "host repo")

	brokenWt := filepath.Join(hostMain, "broken-wt")
	runGitIsolated(t, hostMain, "worktree", "add", "-b", "dev", brokenWt)
	breakWorktreeGitMetadata(t, req, brokenWt)

	req.MainRepo = mainRepo
	req.DepPath = goodChild
	req.ConsumerTop = hostMain
	req.WtDir = brokenWt
	req.RepoDir = mainRepo
}

func breakWorktreeGitMetadata(t *testing.T, req *Request, wtDir string) {
	t.Helper()
	staleGitdir := filepath.Join(req.WorkRoot, "stale-main", ".git", "worktrees", filepath.Base(wtDir))
	writeFile(t, filepath.Join(wtDir, ".git"), "gitdir: "+staleGitdir+"\n")
}

func gitCommandCombinedError(t *testing.T, dir string, args ...string) string {
	t.Helper()
	out, err := git_isolated.Command(dir, args...).CombinedOutput()
	if err == nil {
		t.Fatalf("git %v in %s: expected failure", args, dir)
	}
	return strings.TrimSpace(string(out))
}

func worktreeGitError(t *testing.T, wtPath string) string {
	t.Helper()
	raw := gitCommandCombinedError(t, wtPath, "rev-parse", "--verify", "HEAD")
	return fmt.Sprintf("git rev-parse --verify HEAD in %s: %s", wtPath, raw)
}

func scanBrokenBlockPlain(relDir, statusLine string) string {
	return "Dir:          " + relDir + "\nStatus:       " + statusLine
}

func colorScanRootBlockPlain(t *testing.T, mainRepo string) string {
	t.Helper()
	return fmt.Sprintf("Dir:          .\n%s\n%s\nStatus:       <ansi-color green>clean</ansi-color>\n%s",
		statusBranchLine(t, mainRepo), statusCommitLine(t, mainRepo), statusNoUpstreamRemote())
}

func colorScanStatusBlockPlain(t *testing.T, repoDir, relDir string) string {
	t.Helper()
	return fmt.Sprintf("Dir:          %s\n%s\n%s\nStatus:       <ansi-color green>clean</ansi-color>",
		relDir, statusBranchLine(t, repoDir), statusCommitLine(t, repoDir))
}

func scanErrorStatusPlain(t *testing.T, wtPath string) string {
	t.Helper()
	return "error: " + worktreeGitError(t, wtPath)
}

func scanErrorStatusColored(t *testing.T, wtPath string) string {
	t.Helper()
	gitErr := worktreeGitError(t, wtPath)
	return "<ansi-color red>error: " + gitErr + "</ansi-color>"
}

func assertStdoutBlocksSeparated(t *testing.T, stdout string, wantBlocks int) {
	t.Helper()
	if got := statusOutputBlockCount(stdout); got != wantBlocks {
		t.Fatalf("expected %d status blocks, got %d:\n%s", wantBlocks, got, stdout)
	}
	if wantBlocks > 1 && !strings.Contains(stdout, "\n\n") {
		t.Fatalf("expected blank line between blocks, got:\n%s", stdout)
	}
}

func assertOutputExact(t *testing.T, stdout, template string) {
	t.Helper()
	assert.Output(t, stdout, template)
}

func ensureNestedBrokenLinkedHelpersUsed() {
	_ = setupNestedBrokenLinkedFixture
	_ = breakWorktreeGitMetadata
	_ = scanBrokenBlockPlain
	_ = scanErrorStatusPlain
	_ = scanErrorStatusColored
	_ = colorScanRootBlockPlain
	_ = colorScanStatusBlockPlain
	_ = assertStdoutBlocksSeparated
	_ = assertOutputExact
}
```