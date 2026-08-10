# Scenario

**Feature**: single-repo two modules — free-first apply cascade (C-AP1 + C-AP3 + C-AP4)

```
# root requires shared + local replace; both scopes owned-changed; already main
root → shared (intra-repo)
  -> wrk --unwind --tag-next --push
  -> peel/land prelude (already main; no land)
  -> cascade: tag shared @ pkgs/shared/v0.0.2
  -> pin root require → v0.0.2; KEEP replace => ./pkgs/shared
  -> commit "wrk: cascade pin …"; then tag root @ v0.0.2
  -> push when no pending modules; exit 0
```

## Steps

1. Seed apply single-repo two-module stack (bare origin, clean go.mod Base).
2. Run non-dry-run `--unwind --tag-next --push` (no land flags; already main).
3. Assert free-first tags, keep-replace, cascade pin commit, commit-before-tag.

## Context

- No cross-repo edges → no `--done` / `--merge-back` required.
- **RED** while apply still uses TagNextAll-on-peel without module cascade pin
  commit / keep-replace / commit-before-tag ordering.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyCascadeSingleRepoTwoModules(t, req)
	req.Args = []string{"--unwind", "--tag-next", "--push"}
	return nil
}
```
