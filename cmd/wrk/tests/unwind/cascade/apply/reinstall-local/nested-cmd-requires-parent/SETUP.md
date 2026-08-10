# Scenario

**Bug / gap**: cascade pin nested **cmd/** that depends on **parent** (agent-pro
shape) then `--reinstall-local` — not covered by C-RI1 sibling tools→shared

```
# free monorepo under consumer/external (agent-pro shape):
#   example.com/dot-pkgs                 tagged @ v0.0.2; requires extra@v0.0.2
#   example.com/dot-pkgs/cmd-harness     under cmd/
#     require parent@v0.0.1 + extra@v0.0.1 + replace parent => ../
#     package main cmd/tool
# consumer root: require free@old + droppable replace => ./external/… ; dirty peel
consumer (dirty) + free linked WT (cmd requires parent, replace => ../)
  -> wrk --unwind --tag-next --push --reinstall-local
  -> cascade: pin cmd-harness <- parent @ v0.0.2; KEEP replace => ../
  -> clean free linked Path → pin+tidy on free **main** (reinstall scan root)
  -> reinstall-local scans free main (useMain)
  -> go install ./tool (Dir=cmd) must succeed: failed 0
  -> no "updates to go.mod needed" / missing go.sum
```

## Steps

1. Seed `setupApplyCascadeNestedCmdRequiresParent` (linked free monorepo + consumer).
2. Run `--unwind --tag-next --push --reinstall-local` from consumer.
3. Assert pin keep-replace on free **main** + reinstall without tidy diagnostics.

## Context

- **Production:** `agent-pro` root + nested `cmd/` requires parent with
  `replace => ..`; cascade pin bumps parent require (replace kept); reinstall
  installs `cmd/*` package mains. Stale transitive requires in nested `cmd`
  (e.g. older go-pkgs) need tidy after pin.
- **Fix:** clean free linked Path pins **MainRepo** so useMain reinstall sees pin.
- **Not C-RI1:** sibling `tools` → `pkgs/shared`. This is **child → parent**.
- C-PUSH1 / free multi-module root-only cover cmd+parent pin ordering, **not**
  reinstall after that pin.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyCascadeNestedCmdRequiresParent(t, req)
	req.Args = []string{"--unwind", "--tag-next", "--push", "--reinstall-local"}
	return nil
}
```
