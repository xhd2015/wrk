# Scenario

**Feature**: linked consumer + one `./external/*` dep worktree of another main

```
consumerWt + external/mydep (linked of dep main)
  -> incomplete warm index (consumer only)
  -> wrk --status / --repos from consumerWt
  -> consumer + external must both appear
```

## Steps

1. Linked consumer via `wrk --new`.
2. Dep main + `git worktree add` under `{consumerWt}/external/…`.
3. Ignore `/external` on consumer so porcelain stays clean.
4. Seed incomplete warm cache (consumer-only index) under FakeHome.
5. Leaves choose `--status` or `--repos`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true

	_, consumerWt, _ := setupLinkedConsumer(t, req)

	depMain := setupDepMain(t, req.WorkRoot, "mydep", "dep main for external wt")
	req.DepPath = depMain

	// Bring-shaped name under external/; branch owns dep history.
	relName := "mydep"
	depBranch := branchName("main", wrkDate, 0)
	extDir := addExternalDepWorktree(t, consumerWt, depMain, relName, depBranch)
	req.ExternalWtDir = extDir
	// Reuse Wt2Branch for the dep worktree branch (mega-tree Request has no DepBranch).
	req.Wt2Branch = depBranch

	ignoreExternalOnConsumer(t, consumerWt)

	// Hostile warm: only consumer known under checkout root.
	seedIncompleteConsumerOnlyIndex(t, req.FakeHome, consumerWt)

	req.RepoDir = consumerWt
	return nil
}
```
