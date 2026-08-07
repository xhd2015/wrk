# Scenario

**Feature**: transitive A→B→C via local filesystem replaces; all dirty → free-first C,B,A

```
# A=root, B=agent-pro external sibling, C=dot-pkgs external sibling
# A replace => ../external/agent-pro-main-DATE
# B replace => ../dot-pkgs-main-DATE  (sibling under external/)
# all dirty; pin flags present
transitive local replaces
  -> wrk --unwind --dry-run --tag-next --push
  -> would: peel ../external/dot-pkgs-main-2026-06-30
  -> would: peel ../external/agent-pro-main-2026-06-30
  -> would: peel .
```

## Steps

1. Seed three checkouts linked only by local filesystem replaces (BFS fixpoint).
2. Dirtify all three; dry-run with pin flags.
3. PeelOrder free-first: C then B then A (display paths).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupFollowTransitiveChain(t, req)
	req.Args = []string{"--unwind", "--dry-run", "--tag-next", "--push"}
	recordUnwindBaseline(t, req)
	return nil
}
```
