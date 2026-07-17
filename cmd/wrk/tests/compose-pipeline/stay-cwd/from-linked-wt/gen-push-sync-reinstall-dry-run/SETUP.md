# Scenario

**Feature**: linked wt without done/tag-next — gen-commit + push + sync + reinstall on WT activeRoot

```
# Feature branch worktree; no --tag-next (main-only); no --done
linked wt
  -> wrk --gen-commit-msg --commit --model=m --push --sync --reinstall-local --dry-run
  -> NOT mutually exclusive
  -> gen-commit on WT; push plans feature branch (not forced main-only); reinstall scans WT modules
  -> exit 0; zero mutations
```

## Steps

1. Linked ahead + origin; stage file on wt; seed reinstall on main tip copy into wt via commit on main then?  
   Seed present on main before wt creation is ideal — seed after setup by adding to wt or main.
2. Prefer: setup linked origin, add present on **wt** (activeRoot) for reinstall scan, GOBIN, baseline.
3. Run compose dry-run.

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Setup(t *testing.T, req *Request) error {
	setupAPLinkedAheadOrigin(t, req)

	// Stage uncommitted change for gen-commit on the worktree.
	staged := filepath.Join(req.WtDir, "staged-for-commit.go")
	writeFile(t, staged, "package staged\n")
	runGitIsolated(t, req.WtDir, "add", "staged-for-commit.go")

	// Reinstall scans activeRoot (WT): put cmd/present on the worktree tip + GOBIN stub.
	src := fmt.Sprintf("package %s\n\nfunc main() {}\n", "main")
	cmdDir := filepath.Join(req.WtDir, "cmd", "present")
	if err := os.MkdirAll(cmdDir, 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	writeFile(t, filepath.Join(cmdDir, "main.go"), src)
	runGitIsolated(t, req.WtDir, "add", "cmd/present")
	// Keep staged-for-commit.go staged for gen-commit: unstage it, commit only present.
	runGitIsolated(t, req.WtDir, "restore", "--staged", "staged-for-commit.go")
	runGitIsolated(t, req.WtDir, "commit", "-m", "add present on wt", "--", "cmd/present")
	runGitIsolated(t, req.WtDir, "add", "staged-for-commit.go")

	binDir := filepath.Join(req.WorkRoot, "gobin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatalf("mkdir gobin: %v", err)
	}
	writeFile(t, filepath.Join(binDir, "present"), "stub-binary\n")
	_ = os.Chmod(filepath.Join(binDir, "present"), 0o755)
	req.ExtraEnv = append(req.ExtraEnv, "GOBIN="+binDir)

	recordAPDryRunBaseline(t, req)
	subject := strings.TrimSpace(gitOutputIsolated(t, req.WtDir, "log", "-1", "--format=%s"))
	writeFile(t, filepath.Join(req.WorkRoot, "_compose_dry_run_baseline", "wt.head-subject"), subject+"\n")

	req.Args = []string{
		"--gen-commit-msg", "--commit", "--model=m",
		"--push", "--sync", "--reinstall-local", "--dry-run",
	}
	return nil
}
```
