# Scenario

**Feature**: wrk --bring from inside a linked consumer worktree registers the external worktree under the DEP repo (not the consumer's main repo)

```
# wrk <consumer-main>         -> consumer linked worktree at ~/.wrk/worktrees/consumer-...
# wrk --bring <dep> (from there) -> external/mydep
# external .git gitdir must be <dep-main>/.git/worktrees/..., NOT <consumer-main>/.git/worktrees/...
```

This reproduces the reported case: `wrk $X/agent-pro` creates a consumer
worktree, then `wrk --bring $X/dot-pkgs` run from inside it spawns the dep
worktree. The dep worktree's `.git` was found pointing at
`agent-pro/.git/worktrees/...` (the consumer's main repo) instead of
`dot-pkgs/.git/worktrees/...` (the dep's main repo).

## Preconditions

- Git and Go must be available.
- Consumer main repo has a `go.mod` requiring `example.com/dep`.
- Dep path is a git repo with module `example.com/dep`.

## Steps

1. Create consumer git repo (main checkout) with `go.mod` requiring `example.com/dep`.
2. Create dep git repo `mydep` (main checkout) with module `example.com/dep`.
3. Run `wrk <consumer-main>` to create a linked consumer worktree (mirrors `wrk $X/agent-pro`).
4. Run `wrk --bring <dep>` from inside that linked consumer worktree.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	consumerMain := initBringConsumerRepo(t, req.WorkRoot, true)
	depPath := initBringDepRepo(t, req.WorkRoot, "mydep", true)

	// initBringConsumerRepo leaves go.mod uncommitted; commit it so the linked
	// worktree created next actually checks out a go.mod.
	runGitIsolated(t, consumerMain, "add", "go.mod")
	runGitIsolated(t, consumerMain, "commit", "-m", "add go.mod")

	// Mirror `wrk $X/agent-pro`: spawn a linked worktree of the consumer and
	// run --bring from inside it.
	consumerWt := runWrkFrom(t, req, consumerMain)

	req.RepoDir = consumerWt
	req.DepPath = depPath
	req.ConsumerTop = consumerWt
	req.DepModulePath = bringDepModulePath
	req.Args = []string{"--bring", depPath}
	return nil
}
```
