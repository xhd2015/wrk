# Scenario

**Feature**: apply gen-commit without `--add-all` must not unconditional `git add -A`

```
# linked dep: one staged tracked file + one untracked leftover
# without --add-all: gen-commit commits staged only; untracked stays untracked
staged tracked.txt + untracked leftover.txt
  -> wrk --unwind --gen-commit-msg --commit --merge-back … (no --add-all)
  -> leftover.txt not auto-staged / not in landed commit
```

## Steps

1. Seed linked dep; replace default dirt with:
   - staged tracked change `tracked.txt`
   - untracked `leftover.txt`
2. Run apply gen-commit **without** `--add-all`.
3. Assert leftover is not in dep main HEAD (and preferably still untracked on wt if present).

```go
import (
	"os"
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	seedLinkedDep(t, req)
	// Drop seed untracked change.txt; install controlled staged + untracked pair.
	_ = os.Remove(filepath.Join(req.DepWorktree, "change.txt"))
	req.TrackedName = "tracked.txt"
	req.UntrackedName = "leftover.txt"
	if err := os.WriteFile(filepath.Join(req.DepWorktree, req.TrackedName), []byte("staged body\n"), 0o644); err != nil {
		return err
	}
	git(t, req.DepWorktree, "add", req.TrackedName)
	if err := os.WriteFile(filepath.Join(req.DepWorktree, req.UntrackedName), []byte("leave me\n"), 0o644); err != nil {
		return err
	}
	req.BeforeDep = git(t, req.DepMain, "rev-parse", "HEAD")
	req.BeforeMain = git(t, req.MainRepo, "rev-parse", "HEAD")
	// No --add-all.
	req.Args = unwindGenCommitArgs(t, req, "--merge-back", "--sync", "--tag-next", "--push")
	return nil
}
```
