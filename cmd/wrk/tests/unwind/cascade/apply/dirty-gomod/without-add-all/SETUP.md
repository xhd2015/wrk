# Scenario

**Feature**: dirty go.mod WIP without `--add-all` → partial-edit success (C-AP5 / P3-1)

```
# single-repo two modules; uncommitted go.mod comment WIP; pin needed
root (dirty go.mod vs Base)
  -> wrk --unwind --tag-next --push
  -> save WIP → pin on Base + tidy → cascade pin commit (Base+pin, no WIP)
  -> restore WT: WIP marker preserved + surgical require bump
  -> exit 0
```

## Steps

1. Seed apply single-repo two-module stack (clean Base, bare origin).
2. Append uncommitted WIP comment line to root go.mod (`dirtyRootGoModWIP`).
3. Run without `--add-all`.
4. Expect partial-edit **success** (P3 supersedes P2 hard Error for ordinary WIP).

## Context

- **P2→P3 sealed ASSERT flip (justified):** P2 expected non-zero Error when go.mod
  differed from Base without `--add-all`. Locked product intent **D11**: ordinary
  dirty go.mod WIP uses **partial edit**, not hard fail. Orchestrator pre-approved
  this single ASSERT rewrite for C-AP5 only.
- Pin commit tree must **not** contain the WIP marker; WT must still have it plus
  surgical require version bump and keep-replace.
- Do **not** use `assertGoModCommittedClean` here — WT stays dirty by design.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyCascadeSingleRepoTwoModules(t, req)
	dirtyRootGoModWIP(t, req)
	req.Args = []string{"--unwind", "--tag-next", "--push"}
	return nil
}
```
