# Scenario

**Feature**: linked consumer + two `./external/*` dep worktrees of other mains

```
consumerWt + external/aaa-dep + external/zzz-dep
  + incomplete warm index (consumer only)
  -> wrk --status => 3 Dir blocks
```

## Steps

1. Linked consumer via `wrk --new`.
2. Two dep mains + worktrees under `external/aaa-dep` and `external/zzz-dep`
   (names force path order aaa before zzz).
3. Ignore `/external`; seed incomplete warm cache.
4. Leaf runs `--status`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true

	_, consumerWt, _ := setupLinkedConsumer(t, req)

	depA := setupDepMain(t, req.WorkRoot, "depa", "dep A main")
	depZ := setupDepMain(t, req.WorkRoot, "depz", "dep Z main")
	req.DepPath = depA
	req.DepsDepPath = depZ

	// Stable branch names (ASSERT reconstructs the same literals for Master:).
	branchA := "dep-a-" + wrkDate
	branchZ := "dep-z-" + wrkDate
	extA := addExternalDepWorktree(t, consumerWt, depA, "aaa-dep", branchA)
	extZ := addExternalDepWorktree(t, consumerWt, depZ, "zzz-dep", branchZ)
	req.ExternalWtDir = extA
	req.ExternalWtDir2 = extZ
	req.Wt2Branch = branchA

	ignoreExternalOnConsumer(t, consumerWt)
	seedIncompleteConsumerOnlyIndex(t, req.FakeHome, consumerWt)

	req.RepoDir = consumerWt
	req.Args = []string{"--status"}
	return nil
}
```
