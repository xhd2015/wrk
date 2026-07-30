# Scenario

**Feature**: refuse mutate/delete ops when a branch is checked out in multiple worktrees

```
# shared branch (two linked, or dead second) → hard refuse
git worktree add --force <path2> <branch>
  -> wrk --done | --merge-back | --gen-commit-msg --commit
  -> non-zero; stderr Error: + branch + paths; refuse <op>
  -> no merge / no worktree remove / no branch -D / no new commit

# dead registration still counts as shared
rm -rf path2 (not pruned)
  -> refuse + prune hint: git -C <main> worktree prune

# dry-run fail-closed
wrk --done --dry-run (shared)
  -> non-zero Error: (no plan-as-success)

# unique branch (happy path smoke)
single linked wt on unique branch
  -> wrk --done still succeeds
```

## Preconditions

- Inherits monotree root harness (`Request` / `Response` / `Run`, `setupWrkWorktreeFromMain`,
  `runGitIsolated`, `commitAheadOnWorktree`, worktree asserts).
- Git available; uses `git worktree add --force` to place a second checkout on the same branch.
- Classic TDD **RED** until implementer adds exclusive-branch guards (MergeBack early + wrk
  commit path; format `Error:` with paths / prune hint).
- Data only from gitops `WorktreesOnBranch` (len > 1 including current; dead counted).

## Steps

1. Grouping helpers build shared / unique fixtures on `req`.
2. Leaves set `req.Args`, `req.RepoDir`, and `req.InProcess = true`.

## Context

- Error text flexible but must include: `Error:`, branch name, checkout paths,
  `refuse` + op name (`--done` / `--merge-back` / commit); dead lines show
  `worktree prune` via `git -C <main> …`.
- No force-override flag in product.

```go
import (
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	ensureSharedBranchRefuseHelpersUsed()
	return nil
}

// setupSharedTwoLinked creates a wrk-managed linked worktree and a second linked
// checkout of the same branch via `git worktree add --force`. Commits one ahead
// change on the primary wt so --done would otherwise merge. Sets RepoDir to WtDir.
func setupSharedTwoLinked(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)
	mainRepo, wtDir, branch := setupWrkWorktreeFromMain(t, req)
	mainRepo = compositionResolvePath(t, mainRepo)
	wtDir = compositionResolvePath(t, wtDir)
	req.MainRepo = mainRepo
	req.WtDir = wtDir
	req.WtBranch = branch

	wt2 := filepath.Join(req.WorkRoot, "shared-wt-2")
	runGitIsolated(t, mainRepo, "worktree", "add", "--force", wt2, branch)
	wt2 = compositionResolvePath(t, wt2)
	req.Wt2Dir = wt2
	req.Wt2Branch = branch

	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead of main")
	req.RepoDir = wtDir
}

// setupSharedDead is setupSharedTwoLinked then deletes the second checkout directory
// without pruning (dead/prunable registration remains).
func setupSharedDead(t *testing.T, req *Request) {
	t.Helper()
	setupSharedTwoLinked(t, req)
	if err := os.RemoveAll(req.Wt2Dir); err != nil {
		t.Fatalf("remove second worktree checkout %s: %v", req.Wt2Dir, err)
	}
}

// setupUniqueLinkedAhead is a single wrk-managed linked worktree ahead of main (no share).
func setupUniqueLinkedAhead(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)
	mainRepo, wtDir, branch := setupWrkWorktreeFromMain(t, req)
	mainRepo = compositionResolvePath(t, mainRepo)
	wtDir = compositionResolvePath(t, wtDir)
	req.MainRepo = mainRepo
	req.WtDir = wtDir
	req.WtBranch = branch
	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead of main")
	req.RepoDir = wtDir
}

// setupSharedTwoLinkedStaged stages a text file on the primary shared worktree
// (for --gen-commit-msg --commit). Does not create an ahead commit beyond staging.
func setupSharedTwoLinkedStaged(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)
	mainRepo, wtDir, branch := setupWrkWorktreeFromMain(t, req)
	mainRepo = compositionResolvePath(t, mainRepo)
	wtDir = compositionResolvePath(t, wtDir)
	req.MainRepo = mainRepo
	req.WtDir = wtDir
	req.WtBranch = branch

	wt2 := filepath.Join(req.WorkRoot, "shared-wt-2")
	runGitIsolated(t, mainRepo, "worktree", "add", "--force", wt2, branch)
	wt2 = compositionResolvePath(t, wt2)
	req.Wt2Dir = wt2
	req.Wt2Branch = branch

	writeFile(t, filepath.Join(wtDir, "change.go"), "package main\n")
	runGitIsolated(t, wtDir, "add", "change.go")
	req.RepoDir = wtDir
}

// assertSharedBranchRefuseError checks hard refuse framing for a shared-branch op.
// opHint is a substring naming the operation (e.g. "--done", "--merge-back", "commit").
func assertSharedBranchRefuseError(t *testing.T, req *Request, resp *Response, opHint string) {
	t.Helper()
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for shared-branch refuse (%s); stdout=%q stderr=%q",
			opHint, resp.Stdout, resp.Stderr)
	}
	stderr := resp.Stderr
	if !strings.Contains(stderr, "Error:") {
		t.Fatalf("stderr must include %q; stderr=%q stdout=%q", "Error:", stderr, resp.Stdout)
	}
	if req.WtBranch != "" && !strings.Contains(stderr, req.WtBranch) {
		t.Fatalf("stderr must mention branch %q; stderr=%q", req.WtBranch, stderr)
	}
	// Both checkout paths (or basenames) listed.
	if req.WtDir != "" {
		if !strings.Contains(stderr, req.WtDir) && !strings.Contains(stderr, filepath.Base(req.WtDir)) {
			t.Fatalf("stderr must mention primary worktree path %q; stderr=%q", req.WtDir, stderr)
		}
	}
	if req.Wt2Dir != "" {
		if !strings.Contains(stderr, req.Wt2Dir) && !strings.Contains(stderr, filepath.Base(req.Wt2Dir)) {
			t.Fatalf("stderr must mention second worktree path %q; stderr=%q", req.Wt2Dir, stderr)
		}
	}
	// refuse + operation name
	low := strings.ToLower(stderr)
	if !strings.Contains(low, "refuse") {
		t.Fatalf("stderr must include refuse language; stderr=%q", stderr)
	}
	if opHint != "" && !strings.Contains(low, strings.ToLower(opHint)) {
		t.Fatalf("stderr must name op %q; stderr=%q", opHint, stderr)
	}
}

// assertDeadPruneHint requires dead worktree lines show git worktree prune for main.
func assertDeadPruneHint(t *testing.T, req *Request, resp *Response) {
	t.Helper()
	stderr := resp.Stderr
	if !strings.Contains(stderr, "worktree prune") {
		t.Fatalf("dead shared refuse must mention worktree prune; stderr=%q", stderr)
	}
	if !strings.Contains(stderr, "git") || !strings.Contains(stderr, "-C") {
		t.Fatalf("prune hint should look like git -C <main> worktree prune; stderr=%q", stderr)
	}
	if req.MainRepo != "" {
		if !strings.Contains(stderr, req.MainRepo) && !strings.Contains(stderr, filepath.Base(req.MainRepo)) {
			t.Fatalf("prune hint should reference main repo %q; stderr=%q", req.MainRepo, stderr)
		}
	}
}

// assertNoDoneMutations: primary wt, second wt (if live), and branch remain.
func assertNoDoneMutations(t *testing.T, req *Request) {
	t.Helper()
	assertFileExists(t, req.WtDir)
	assertWorktreeListContains(t, req.MainRepo, req.WtDir)
	assertBranchExists(t, req.MainRepo, req.WtBranch)
	// feature-work not on main (no merge)
	assertFileNotExists(t, filepath.Join(req.MainRepo, "feature-work"))
	if req.Wt2Dir != "" {
		if _, err := os.Stat(req.Wt2Dir); err == nil {
			assertWorktreeListContains(t, req.MainRepo, req.Wt2Dir)
		}
	}
}

func ensureSharedBranchRefuseHelpersUsed() {
	_ = setupSharedTwoLinked
	_ = setupSharedDead
	_ = setupUniqueLinkedAhead
	_ = setupSharedTwoLinkedStaged
	_ = assertSharedBranchRefuseError
	_ = assertDeadPruneHint
	_ = assertNoDoneMutations
}
```
