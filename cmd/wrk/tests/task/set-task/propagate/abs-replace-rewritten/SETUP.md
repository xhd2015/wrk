# Scenario

**Feature**: --set-task rewrites absolute go.mod replace paths after consumer rename

```
# wrk --dep writes replace example.com/dep => <abs>/external/mydep-...
# --set-task moves consumer wt → go.mod replace must point at new external path
consumer wt + wrk --dep (abs replace) -> wrk --set-task -> go.mod replace updated
```

## Steps

1. Create consumer main + dep, spawn consumer wt with `--task "old slug"`.
2. Run `wrk --dep` from consumer wt (writes absolute replace in go.mod).
3. Record old consumer wt path and external wt path.
4. Run `wrk --set-task "new slug"` with `WRK_SET_TASK_CONFIRM=1`.

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

	runGitIsolated(t, consumerMain, "add", "go.mod")
	runGitIsolated(t, consumerMain, "commit", "-m", "add go.mod")

	consumerWt := runWrkWithArgs(t, req, consumerMain, "--task", "old slug")
	req.WtDir = consumerWt
	req.MainRepo = consumerMain
	req.ConsumerTop = consumerWt

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
	req.ExternalWtDir = strings.TrimSpace(string(out))
	req.DepPath = dep

	// Precondition: wrk --dep wrote an absolute replace under the old consumer path.
	pre, err := propagateReadGoMod(consumerWt)
	if err != nil {
		t.Fatalf("read go.mod before set-task: %v", err)
	}
	oldReplace := propagateReplacePathForModule(pre, depModulePath)
	if oldReplace == "" {
		t.Fatalf("expected replace for %s before set-task", depModulePath)
	}
	if !filepath.IsAbs(oldReplace) {
		t.Fatalf("wrk --dep should write absolute replace, got %q", oldReplace)
	}
	if !strings.HasPrefix(oldReplace, consumerWt) {
		t.Fatalf("abs replace should be under consumer wt %q, got %q", consumerWt, oldReplace)
	}

	req.RepoDir = consumerWt
	req.SetTaskDesc = "new slug"
	req.SetTaskEnv = "WRK_SET_TASK_CONFIRM=1"
	return nil
}
```