# Scenario

**Feature**: --set-task updates gitdir metadata for a single external dep worktree

```
# consumer wt with one external dep → rename → dep's gitdir updated to new path
wrk --set-task "new slug" (WRK_SET_TASK_CONFIRM=1, from consumer wt with external dep)
  -> discovers dep under external/
  -> git worktree move (consumer)
  -> updates <depMain>/.git/worktrees/<name>/gitdir to new external path
```

## Steps

1. Create consumer main repo with go.mod requiring a dep.
2. Create dep main repo.
3. Spawn consumer linked worktree with `--task "old slug"`.
4. Run `wrk --dep <dep>` from inside consumer worktree.
5. Store old external path and old gitdir content.
6. Run `wrk --set-task "new slug"` with `WRK_SET_TASK_CONFIRM=1`.
7. Verify dep's gitdir in dep main repo now points to new path.

```go
import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func Setup(t *testing.T, req *Request) error {
	consumerMain := initConsumerRepo(t, req.WorkRoot, true)
	dep := initDepRepo(t, req.WorkRoot, "mydep")

	// Commit consumer go.mod so linked worktree can be created
	runGitIsolated(t, consumerMain, "add", "go.mod")
	runGitIsolated(t, consumerMain, "commit", "-m", "add go.mod")

	// Spawn consumer worktree with --task
	consumerWt := runWrkWithArgs(t, req, consumerMain, "--task", "old slug")
	req.WtDir = consumerWt
	req.MainRepo = consumerMain

	// Run --dep from inside consumer worktree
	depCmd := exec.Command(getWrkBin(t), "--dep", dep)
	depCmd.Dir = consumerWt
	depCmd.Env = append(os.Environ(),
		"WRK_HOME="+req.WrkHome,
		"WRK_DATE="+wrkDate,
	)
	out, err := depCmd.Output()
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			t.Fatalf("wrk --dep exit %d stderr=%q", ee.ExitCode(), string(ee.Stderr))
		}
		t.Fatalf("wrk --dep: %v", err)
	}
	extPath := strings.TrimSpace(string(out))
	req.ExternalWtDir = extPath

	// Store old gitdir content for comparison
	oldGitdir, err := readWorktreeGitdir(extPath)
	if err != nil {
		t.Fatalf("read old gitdir: %v", err)
	}
	req.OldExternalGitdir = oldGitdir

	req.RepoDir = consumerWt
	req.SetTaskDesc = "new slug"
	req.SetTaskEnv = "WRK_SET_TASK_CONFIRM=1"
	req.DepPath = dep
	req.ConsumerTop = consumerWt
	return nil
}
```
