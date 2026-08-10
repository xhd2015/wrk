# Scenario

**Feature**: multi-repo free-first cascade + reinstall-local tail (C-RI2)

```
# root requires leaf@v0.0.1; both dirty; leaf owned-changed → next v0.0.2
leaf ← root (repos + modules)
  -> wrk --unwind --tag-next --push --done --reinstall-local
  -> land free-first + cascade tag/pin as C-AP2
  -> reinstall-local on collected mains; exit 0
  -> no unknown revision; no go mod tidy hard failure
```

## Steps

1. Seed multi-repo apply cascade fixture (both dirty; bare origins; modproxy).
2. Isolate GOBIN (no required install candidates — empty reinstall OK).
3. Run with `--reinstall-local` in addition to C-AP2 flags.
4. Assert cascade pin still holds and reinstall does not hard-fail.

## Context

- Reuses `setupApplyCascadeMultiRepoBothDirty` (light fixture; optional matrix).
- Mixed mode OK: may already be **GREEN** after P2 cascade if reinstall tail is
  soft; still required as regression that the full compose path exits 0.
- Keep-replace remains single-repo nested-skip leaf (C-RI1); multi-repo uses
  modproxy pin without local replace on root.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyCascadeMultiRepoBothDirty(t, req)
	// Isolated GOBIN; no bin stubs → reinstall plan may be empty (soft OK).
	enableIsolatedReinstallGOBIN(t, req, "")
	req.Args = []string{"--unwind", "--tag-next", "--push", "--done", "--reinstall-local"}
	return nil
}
```
