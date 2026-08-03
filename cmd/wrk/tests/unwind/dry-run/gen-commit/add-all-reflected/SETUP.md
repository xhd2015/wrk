# Scenario

**Feature**: gen-commit + `--add-all` dry-run prints `would: git add -A` under peel

```
# sole dirty main; untracked DIRTY present
root (dirty) -> wrk --unwind --dry-run --gen-commit-msg --commit --add-all
  -> would: peel .
  ->   would: git add -A
  ->   would: generate commit message and commit staged changes
  -> exit 0; zero mutations
```

## Steps

1. Seed single dirty main (untracked `DIRTY`).
2. Run unwind dry-run with gen-commit + `--add-all` + `--commit`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupSingleMainDirty(t, req)
	req.Args = []string{
		"--unwind", "--dry-run",
		"--gen-commit-msg", "--commit", "--add-all",
	}
	recordUnwindBaseline(t, req)
	return nil
}
```
