# Scenario

**Feature**: TTY with multiple reusable same-parent siblings skips to lex-smallest; also-present for others

```
# two clean same-HEAD siblings under target: worktree-A < worktree-B (lex)
# TTY + Enter -> skip; stdout = worktree-A; also present: worktree-B
worktree-A + worktree-B under {WorkRoot}/target
  -> wrk myrepo {WorkRoot}/target (\n)
  -> stdout smallest; no new under target
```

## Steps

1. Pre-create two reusable siblings `worktree-A` and `worktree-B` under target.
2. Run named create under fake TTY with stdin `\n`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	parent := policyBTargetParent(req)
	// Names chosen so abs path lex order is A then B.
	a := policyBAddSiblingUnderParent(t, req, parent, "worktree-A", "policy-b-reuse-a")
	b := policyBAddSiblingUnderParent(t, req, parent, "worktree-B", "policy-b-reuse-b")
	if a >= b {
		t.Fatalf("fixture lex order: want %q < %q", a, b)
	}
	req.WtDir = a            // primary (lex-smallest)
	req.ExternalWtDir2 = b   // second reusable path
	req.StdinInput = "\n"
	return nil
}
```
