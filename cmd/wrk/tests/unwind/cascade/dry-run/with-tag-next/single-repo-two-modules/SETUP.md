# Scenario

**Feature**: single-repo two modules — free-first tag leaf then pin root (C-DR1)

```
# root requires shared; shared owned-changed; single dirty main
root → shared (intra-repo modules)
  -> wrk --unwind --dry-run --tag-next
  -> would: peel .
  -> would: tag-next example.com/root/shared @ pkgs/shared/v0.0.2
  -> would: pin example.com/root <- example.com/root/shared @ v0.0.2
  -> exit 0; zero mutations
```

## Steps

1. Seed single dirty main with root + `pkgs/shared` (owned change + tags).
2. Run dry-run with `--tag-next` only (no cross-repo edges → no `--push` / land).
3. Expect peel `.` then free-first cascade: tag shared before pin root.

## Context

- No cross-repo edges → pin/land flag validation does not require `--push`/`--done`.
- Module edge: root depends on shared → free-first tags shared first.
- **RED** until cascade plan prints `would: tag-next` / cascade pin lines.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupCascadeSingleRepoTwoModules(t, req)
	// Single-repo: no cross-repo edges; --tag-next alone enables cascade plan.
	req.Args = []string{"--unwind", "--dry-run", "--tag-next"}
	recordUnwindBaseline(t, req)
	return nil
}
```
