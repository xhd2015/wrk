# Scenario

**Feature**: TTY reusable sibling + default Enter skips create; stdout = sibling path

```
# clean same-HEAD sibling worktree-A under target
# TTY + Enter (default Y) -> skip; stdout = sibling; no new under target
{WorkRoot}/target/worktree-A (reusable)
  -> wrk myrepo {WorkRoot}/target  (stdin \n)
  -> stdout worktree-A; no target/myrepo-main-{date}
```

## Steps

1. Add reusable sibling `worktree-A` under `{WorkRoot}/target` (clean, HEAD==source).
2. Run named create under fake TTY with stdin `\n`.

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	parent := policyBTargetParent(req)
	sib := policyBAddSiblingUnderParent(t, req, parent, "worktree-A", "policy-b-reuse-sib")
	// Ensure clean + same HEAD as source.
	srcHEAD := revParseHEAD(t, req.TargetDir)
	if got := revParseHEAD(t, sib); got != srcHEAD {
		t.Fatalf("sibling HEAD %s != source %s", got, srcHEAD)
	}
	if st := strings.TrimSpace(gitOutputIsolated(t, sib, "status", "--porcelain")); st != "" {
		t.Fatalf("sibling not clean: %q", st)
	}
	req.WtDir = sib
	req.StdinInput = "\n"
	return nil
}
```
