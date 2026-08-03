# Scenario

**Feature**: non-dry-run `--unwind` apply peels free-first with explicit ship/land flags

```
# stack apply (no --dry-run)
dirty pending free-first
  -> cycle/flag validation before any mutation
  -> peel each free dirty repo: optional commit → land (if linked) → tag/push/…
  -> after peeling dep U: Pin consumers at in-scope Path + go mod tidy
  -> --done removes linked WT; reinstall soft on fail
```

## Preconditions

- Parent helpers: `setupApplyLeafPinStack`, `setupApplyLinkedConsumerPinStack`,
  `setupApplyAlreadyMainRootBump`, pin/tag/origin asserts, local bare remotes,
  optional `file://` modproxy.
- Leaves set `req.InProcess = true` and full `req.Args` **without** `--dry-run`.
- Pin target Path ≠ MainRepo leaf expects **RED** until implementer stops remapping
  consumers to MainRepo; pin-when-primary-is-main (leaf-then-pin) stays **GREEN**.

## Steps

1. Grouping marks the apply family; leaves seed fixtures and flags.
2. All leaves under this node run non-dry-run `--unwind` (mutate on success).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = t
	req.InProcess = true
	unwindEnsureHelpersUsed()
	return nil
}
```
