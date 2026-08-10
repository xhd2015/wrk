# Scenario

**Feature**: testdata / forever-skip scopes never get cascade tag-next (C-DR4)

```
# real free module shared taggable; testdata/x also owned-changed + path tags
root + shared + testdata/x
  -> wrk --unwind --dry-run --tag-next
  -> would: tag-next example.com/root/shared @ …
  -> MUST NOT would: tag-next …testdata… / example.com/root/testdata-x
  -> exit 0; zero mutations
```

## Steps

1. Seed single-repo two modules **plus** `testdata/x` go.mod and path-scope tags.
2. Run dry-run with `--tag-next`.
3. Positive: shared still free-first tagged; negative: no testdata tag-next.

## Context

- `scan.Scan` prunes `testdata/` go.mod; tagscope path scopes under `testdata/`
  must still be excluded from cascade **tag** lines if planned.
- Positive shared tag-next keeps this leaf **RED** on current product (no cascade).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupCascadeSingleRepoTwoModulesWithTestdata(t, req)
	req.Args = []string{"--unwind", "--dry-run", "--tag-next"}
	recordUnwindBaseline(t, req)
	return nil
}
```
