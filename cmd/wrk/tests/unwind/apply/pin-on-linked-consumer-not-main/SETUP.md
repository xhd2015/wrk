# Scenario

**Feature**: pin-on-linked-consumer-not-main — pin Path (linked WT), not MainRepo

```
# consumer main (outside scope): require dep@v0.0.1, no replace — baseline snapshotted
# consumer linked WT (primary Path): require + replace → external/leaf; RepoDir = WT
# nested leaf external: dirty + ahead; bare origin on leaf main
linked consumer Path + nested leaf
  -> wrk --unwind --done --tag-next --push   (from linked WT)
  -> peel leaf: land → tag v0.0.2 → push
  -> Pin Path go.mod → require v0.0.2, drop replace
  -> consumer MainRepo go.mod unchanged (still v0.0.1, no pin)
```

## Steps

1. Build linked-consumer apply stack (`setupApplyLinkedConsumerPinStack`):
   materialize **both** consumer main baseline and consumer linked WT with stack replace.
2. Run non-dry-run unwind with land + pin flags from **linked WT** (RepoDir = WT).

## Context

- Expect **RED** while product remaps pin consumers to `MainRepo` (main mutated, WT keep replace).
- GREEN when pin uses in-scope StackMember.Path only.
- Pin log may still say `pin root <- dot-pkgs @ v…` (basename human); assert **file** surfaces.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupApplyLinkedConsumerPinStack(t, req)
	// Flag order free; --done lands linked leaf; pin via tag-next+push.
	req.Args = []string{"--unwind", "--done", "--tag-next", "--push"}
	return nil
}
```
