# Scenario

**Feature**: `wrk --all-deps --dry-run` plans reuse of an existing external dep worktree without writes

```
# existing external/mydep1-main-{date} (no replace) + registered dep
# --all-deps --dry-run -> would: line uses existing name; reuse warning; no new WT/replace/gitignore
```

## Steps

1. Create and register `mydep1`; consumer requires dep1 (no replace).
2. Manually add a live external worktree at the preferred path (same as live reuse leaf).
3. Run `wrk --all-deps --dry-run` via `Run`.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	allDepsEnsureHelpersUsed()
	dryRunEnsureHelpersUsed()

	dep1 := allDepsDepDir(req.WorkRoot, "mydep1")
	initAllDepsRepo(t, dep1, "example.com/dep1", "dep1")
	registerAllDepsProjects(t, req, dep1)

	consumer := initAllDepsConsumer(t, req.WorkRoot, []string{"example.com/dep1"}, "")

	wantExt := allDepsExternalAbsPath(consumer, "mydep1")
	mkdirAll(t, filepath.Join(consumer, "external"))
	runGitIsolated(t, dep1, "worktree", "add", "-b", branchName("main", wrkDate, 0), wantExt)

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.DepPath = dep1
	req.ExternalWtDir = wantExt
	req.Args = []string{"--all-deps", "--dry-run"}
	return nil
}
```
