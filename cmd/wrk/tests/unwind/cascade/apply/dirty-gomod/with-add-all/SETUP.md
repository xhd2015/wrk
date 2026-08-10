# Scenario

**Feature**: dirty go.mod WIP **with** `--add-all` succeeds (C-AP6)

```
# same WIP go.mod dirt as C-AP5; --add-all allows cascade pin staging
root (dirty go.mod vs Base)
  -> wrk --unwind --tag-next --push --add-all
  -> cascade pin may stage go.mod/go.sum (+ add-all extras)
  -> commit "wrk: cascade pin …"; tags; exit 0
```

## Steps

1. Seed apply single-repo two-module stack.
2. Append uncommitted WIP line to root go.mod.
3. Run with `--add-all` (cascade staging companion; not dry-run).
4. Expect success: pin commit, require bump, keep replace, tags.

## Context

- Today top-level `--add-all` may reject without `--commit` — **RED** until
  implementer accepts `--add-all` with `--unwind` cascade (or equivalent wiring
  into `UnwindFlags`).
- Selective pin commit may include the WIP line when `--add-all` stages it.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyCascadeSingleRepoTwoModules(t, req)
	dirtyRootGoModWIP(t, req)
	// --add-all with --unwind: cascade may stage extra WIP when pin dirties Base.
	req.Args = []string{"--unwind", "--tag-next", "--push", "--add-all"}
	return nil
}
```
