# Scenario

**Feature**: C-PUSH1 — nested same-main cascade pin must not defer free push
past cross-repo network pin of free @ next

```
# free: example.com/dot-pkgs + example.com/dot-pkgs/cmd (same git main)
# consumer: example.com/app (path < nested cmd → Kahn free order pins app first)
# production analogue: agent-pro tag, cmd-doctest-harness same-main pending,
#   spl (earlier alpha path) pin + go mod tidy needs published tag
free multi-module + app consumer
  -> wrk --unwind --tag-next --push --done
  -> tag-next free root @ v0.0.2
  -> pushed free main (branch+tags)  # must precede cross-repo pin
  -> pin app <- dot-pkgs @ v0.0.2
  -> exit 0; consumer require free @ v0.0.2
```

## Steps

1. Seed C-PUSH1 fixture (`setupApplyCascadePushBeforeCrossRepoPin`).
2. Run apply with land + `--tag-next --push`.
3. Assert free push appears after free root tag-next and **before** cross-repo
   pin of free @ next (order RED under `remainingTouchesMain` holding free
   push for nested same-main work).

## Context

- **Bug:** `maybePushMain` skips free push while later steps touch free main
  (nested `cmd` pin). Kahn free queue sorts `example.com/app` before
  `example.com/dot-pkgs/cmd`, so cross-repo pin runs first → production
  `unknown revision` when next is not yet on remote; this leaf locks push
  order with offline modproxy still seeding next (order-only RED).
- Classic TDD: **RED** until free is published before network pin of free.
- Do not rewrite sealed multi-repo free-first / C1 free-multimodule asserts.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyCascadePushBeforeCrossRepoPin(t, req)
	req.Args = []string{"--unwind", "--tag-next", "--push", "--done"}
	return nil
}
```
