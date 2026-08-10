# Scenario

**Feature**: nested skip consumer pinned then reinstall-local succeeds (C-RI1)

```
# monorepo: pkgs/shared owned-changed; tools/ skip-tag consumer with old require + replace
shared (owned-changed) <- tools (skip tag; require@old + replace)
  -> wrk --unwind --tag-next --push --reinstall-local
  -> cascade: tag shared @ pkgs/shared/v0.0.2
  -> pin tools require → v0.0.2; KEEP replace => ../pkgs/shared
  -> no tag-next for tools
  -> reinstall-local: go install tools/cmd/tool; failed 0
  -> no unknown revision; no go mod tidy failure
```

## Steps

1. Seed nested-skip monorepo (shared owned-changed; tools skip consumer + GOBIN stub).
2. Run apply cascade with `--reinstall-local` (already main; no land flags).
3. Assert tools require bump + keep-replace, shared tag, reinstall success.

## Context

- **Original failure mode:** nested module previously “skip” still requiring old
  dep → reinstall/tidy surfaces `unknown revision` or go-mod-tidy diagnostics.
- **After cascade:** tools require matches new tag; replace kept; reinstall exit 0.
- Classic TDD: **RED** if product still fails nested reinstall after cascade;
  **GREEN** backfill OK if P2/P3 already fixed it (still keep the leaf).
- Dry-run vocabulary for reinstall tail remains **C-DR6** (sealed).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyCascadeNestedSkipConsumer(t, req)
	req.Args = []string{"--unwind", "--tag-next", "--push", "--reinstall-local"}
	return nil
}
```
