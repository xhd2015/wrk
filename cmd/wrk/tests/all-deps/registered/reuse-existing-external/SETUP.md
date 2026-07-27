# Scenario

**Feature**: `wrk --all-deps` reuses an existing live external worktree of a registered dep main

```
# manual external/mydep1-main-{date} of depMain (no go.mod replace yet)
# --all-deps reuses that path, applies replace, no second WT; reuse warning on stderr
consumer (require dep1) + projects.json(mydep1)
  + existing external wt of mydep1 (no replace)
  -> wrk --all-deps
  -> stdout wrk line uses existing ./external/mydep1-main-{date}
  -> replace applied; no -1 dir
```

## Steps

1. Create and register `mydep1` (`example.com/dep1`).
2. Create consumer requiring dep1 (no pre-existing replace).
3. Manually add a live linked worktree of `mydep1` at the preferred external path (no `go.mod` replace — avoids the already-replaced skip and avoids `go mod tidy` dropping unused requires).
4. Run `wrk --all-deps` via `Run` — must reuse that path and apply replace.

```go
import (
	"github.com/xhd2015/doctest/session"
	"path/filepath"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	allDepsEnsureHelpersUsed()
	registeredEnsureHelpersUsed()

	dep1 := allDepsDepDir(req.WorkRoot, "mydep1")
	initAllDepsRepo(t, dep1, "example.com/dep1", "dep1")
	registerAllDepsProjects(t, req, dep1)

	consumer := initAllDepsConsumer(t, req.WorkRoot, []string{"example.com/dep1"}, "")

	// Pre-seed live external WT of the same depMain (no replace in go.mod).
	wantExt := allDepsExternalAbsPath(consumer, "mydep1")
	mkdirAll(t, filepath.Join(consumer, "external"))
	runGitIsolated(t, dep1, "worktree", "add", "-b", branchName("main", wrkDate, 0), wantExt)

	req.RepoDir = consumer
	req.ConsumerTop = consumer
	req.DepPath = dep1
	req.ExternalWtDir = wantExt
	req.Args = []string{"--all-deps"}
	return nil
}
```
