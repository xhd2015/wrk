# Scenario

**Feature**: TTY reusable sibling + `n` creates a new worktree under target

```
# reusable sibling remains; user answers n
# create proceeds as today: named subdir under existing target dir
{WorkRoot}/target/worktree-A (reusable)
  -> wrk myrepo {WorkRoot}/target  (stdin n\n)
  -> new {WorkRoot}/target/myrepo-main-{date}
```

## Steps

1. Add reusable sibling `worktree-A` under target.
2. Run named create under fake TTY with stdin `n\n`.

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	parent := policyBTargetParent(req)
	sib := policyBAddSiblingUnderParent(t, req, parent, "worktree-A", "policy-b-reuse-sib-n")
	req.WtDir = sib
	req.StdinInput = "n\n"
	return nil
}
```
