# Scenario

**Feature**: full ship from linked worktree: gen-commit → done → sync → tag-next → push → reinstall (dry-run; activeRoot switches to main after done)

```
# Linked wtA ahead + staged; wtB; origin; GOBIN present stub
linked wt
  -> wrk --gen-commit-msg --commit --model=m --done --sync --tag-next --push --reinstall-local --dry-run
  -> stage 1 gen-commit plan on WT (activeRoot=WT)
  -> stage 2 done MergeBack DryRun plan (remove)
  -> activeRoot would be main for stages 3–7
  -> sync → tag-next → push → reinstall plans in that order
  -> exit 0; zero mutations
```

## Steps

1. Sync+origin fixture; stage uncommitted file on wt; seed reinstall; baseline.
2. Run full ship with `--dry-run` (no `-y`).

```go
import (
	"path/filepath"
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupAPSyncWithOrigin(t, req)

	staged := filepath.Join(req.WtDir, "staged-for-commit.go")
	writeFile(t, staged, "package staged\n")
	runGitIsolated(t, req.WtDir, "add", "staged-for-commit.go")

	_ = seedAPReinstallPresent(t, req)
	recordAPDryRunBaseline(t, req)

	subject := strings.TrimSpace(gitOutputIsolated(t, req.WtDir, "log", "-1", "--format=%s"))
	writeFile(t, filepath.Join(req.WorkRoot, "_compose_dry_run_baseline", "wt.head-subject"), subject+"\n")

	req.Args = []string{
		"--gen-commit-msg", "--commit", "--model=m",
		"--done", "--sync", "--tag-next", "--push",
		"--reinstall-local", "--dry-run",
	}
	return nil
}
```
