# Scenario

**Feature**: wrk --gen-commit-msg --add-all forwards / peels for bare + compose paths

```
# bare: --add-all is a library flag; wrk path must forward it (not reject as unknown)
git repo with staged files
  -> wrk --gen-commit-msg --add-all --dry-run
  -> stderr: would: git add -A
  -> not: unrecognized flag / unknown flag for --add-all

# compose: peel --add-all with primary so lessflags never sees it
linked wt (ahead) + staged
  -> wrk --gen-commit-msg --add-all --commit --done --dry-run
  -> not: unrecognized flag: --add-all
  -> stderr: would: git add -A
  -> gen-commit dry plan + primary dry plan (exit 0)
```

## Preconditions

- Root harness has initialized WorkRoot / WrkHome.
- Leaves stage a repo (or linked worktree) and set Args including `--add-all`.

## Steps

1. Inherit root Setup.
2. Leaf stages files / linked wt and sets `--gen-commit-msg --add-all …`.

## Context

- Library dry-run prints `would: git add -A\n` on stderr when `--add-all` is set.
- Bare path strips only `--gen-commit-msg` and forwards remaining flags to
  `commit_msg.RunGenCommitMsg` (no wrk-side peel required for bare mode).
- Compose with `--done` / `--merge-back`: `peelGenCommitMsgForCompose` must peel
  `--add-all` via `genCommitMsgBoolFlags` (like `--commit` / `--no-verify`);
  otherwise lessflags reports `unrecognized flag: --add-all` (**Classic RED**).
- Must not confuse with wrk project `--add` (disallowed with bare gen-commit-msg).

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	// Grouping: add-all leaves share stage + compose helpers.
	ensureGenCommitMsgHelpersUsed()
	_ = setupLinkedWtWithStagedForCompose
	return nil
}

// setupLinkedWtWithStagedForCompose builds main + linked worktree ahead with one
// staged text file. RepoDir is the worktree (process cwd for wrk --done).
// Staged-only dirt keeps MergeBack dry-run viable after staged stash.
func setupLinkedWtWithStagedForCompose(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)

	main := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepo(t, main)
	req.MainRepo = main

	wt := filepath.Join(req.WorkRoot, "wt")
	branch := "feature-work"
	runGit(t, main, "worktree", "add", "-b", branch, wt)
	req.WtBranch = branch

	// Commit ahead so MergeBack dry-run plans ff-merge + remove + branch -D.
	writeFile(t, filepath.Join(wt, "feature.go"), "package feature\n")
	runGit(t, wt, "add", "feature.go")
	runGit(t, wt, "commit", "-m", "feature work")

	// Staged change for gen-commit dry plan (+ --add-all would-line still prints).
	writeFile(t, filepath.Join(wt, "staged-for-commit.go"), "package staged\n")
	runGit(t, wt, "add", "staged-for-commit.go")

	req.RepoDir = wt
	req.HEADSubject = gitHEADSubject(t, wt)
}
```
