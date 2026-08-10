# Scenario

**Feature**: partial edit with dirty go.mod + unrelated WIP file (P3-2)

```
# go.mod comment WIP + untracked WIP_NOTES.md; no --add-all
root dirty mods + unrelated file
  -> wrk --unwind --tag-next --push
  -> cascade pin commit only go.mod/go.sum
  -> WIP_NOTES.md still untracked; go.mod WIP preserved + surgical bump
```

## Steps

1. Seed apply single-repo two-module stack.
2. Dirty root go.mod WIP; add untracked `WIP_NOTES.md`.
3. Run without `--add-all`.
4. Assert selective pin commit + unrelated file untouched.

## Context

- Selective cascade commit must not stage non-module WIP even when go.mod is dirty.
- **RED** until partial edit + selective commit.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyCascadeSingleRepoTwoModules(t, req)
	dirtyRootGoModWIP(t, req)
	dirtyUnrelatedWIPFile(t, req)
	req.Args = []string{"--unwind", "--tag-next", "--push"}
	return nil
}
```
