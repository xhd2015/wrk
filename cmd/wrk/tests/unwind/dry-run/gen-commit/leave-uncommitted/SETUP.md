# Scenario

**Feature**: gen-commit without `--add-all` plans leave-N for not-fully-staged paths

```
# sole dirty main; untracked DIRTY (N=1 not-fully-staged)
root (dirty) -> wrk --unwind --dry-run --gen-commit-msg --commit
  -> would: peel .
  ->   would: leave 1 file uncommitted (use --add-all if necessary)
  ->   would: generate commit message and commit staged changes
  -> exit 0; zero mutations
```

## Steps

1. Seed single dirty main with one untracked `DIRTY` file (N=1).
2. Run unwind dry-run with gen-commit + commit; **omit** `--add-all`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupSingleMainDirty(t, req)
	req.LeaveN = 1
	req.Args = []string{
		"--unwind", "--dry-run",
		"--gen-commit-msg", "--commit",
	}
	recordUnwindBaseline(t, req)
	return nil
}
```
