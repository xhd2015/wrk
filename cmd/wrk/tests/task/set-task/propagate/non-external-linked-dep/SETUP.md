# Scenario

**Feature**: --set-task updates gitdir metadata for a manual linked dep worktree outside external/

```
# consumer wt with deps/foo linked wt (manual git worktree add) → rename → dep gitdir updated
wrk --set-task "new slug" (WRK_SET_TASK_CONFIRM=1, from consumer wt with deps/foo)
  -> discovers deps/foo via scan_repo (not under external/)
  -> git worktree move (consumer)
  -> updates <depMain>/.git/worktrees/<name>/gitdir to new deps/foo path
```

## Steps

1. Create consumer main repo with go.mod.
2. Create dep main repo.
3. Spawn consumer linked worktree with `--task "old slug"`.
4. Run `git -C <depMain> worktree add -b <branch> {consumerWt}/deps/foo` (manual linked wt, NOT via `--dep`).
5. Store old gitdir for deps/foo.
6. Run `wrk --set-task "new slug"` with `WRK_SET_TASK_CONFIRM=1`.
7. Verify dep's gitdir in dep main repo now points to new path.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.InProcess = true
	_ = d
	consumerMain := initConsumerRepo(t, req.WorkRoot, false)

	dep := initDepRepo(t, req.WorkRoot, "foodep")

	runGitIsolated(t, consumerMain, "add", "go.mod")
	runGitIsolated(t, consumerMain, "commit", "-m", "add go.mod")

	consumerWt := runWrkWithArgs(t, req, consumerMain, "--task", "old slug")
	req.WtDir = consumerWt
	req.MainRepo = consumerMain

	depsWtDir := filepath.Join(consumerWt, "deps", "foo")
	req.DepsLinkedWtDir = depsWtDir
	req.DepsDepPath = dep
	req.DepPath = dep

	depBranch := branchName("main", wrkDate, 0)
	runGitIsolated(t, dep, "worktree", "add", "-b", depBranch, depsWtDir)

	writeFile(t, filepath.Join(consumerWt, ".gitignore"), "/deps\n")
	runGitIsolated(t, consumerWt, "add", ".gitignore")
	runGitIsolated(t, consumerWt, "commit", "-m", "ignore deps worktrees")

	oldGitdir, err := readWorktreeGitdir(depsWtDir)
	if err != nil {
		t.Fatalf("read old gitdir: %v", err)
	}
	req.OldExternalGitdir = oldGitdir

	req.RepoDir = consumerWt
	req.SetTaskDesc = "new slug"
	req.SetTaskEnv = "WRK_SET_TASK_CONFIRM=1"
	return nil
}
```