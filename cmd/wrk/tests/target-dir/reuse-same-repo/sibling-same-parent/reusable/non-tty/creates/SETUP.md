# Scenario

**Feature**: non-TTY + reusable same-parent sibling creates a new worktree (automation-friendly)

```
# reusable sibling worktree-A under target; non-TTY
# exit 0; stdout = new target/myrepo-main-{date}; no refuse wording
# no skip prompt (non-interactive create, not refuse)
{WorkRoot}/target/worktree-A (reusable) + non-TTY
  -> wrk myrepo {WorkRoot}/target
  -> create new path
```

## Steps

1. Pre-create one reusable sibling under `{WorkRoot}/target`.
2. Run `wrk myrepo {WorkRoot}/target` without TTY (default doctest pipe, InProcess).

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	parent := policyBTargetParent(req)
	sib := policyBAddSiblingUnderParent(t, req, parent, "worktree-A", "policy-b-reuse-nontty")
	req.WtDir = sib
	return nil
}
```
