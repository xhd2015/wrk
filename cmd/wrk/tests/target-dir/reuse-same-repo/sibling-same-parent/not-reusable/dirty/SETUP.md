# Scenario

**Feature**: dirty same-parent sibling is not reusable → create without Policy B banner

```
# sibling worktree-A under target has untracked/dirty porcelain
# wrk myrepo target -> create myrepo-main-{date}; no skip prompt
myrepo + dirty sibling {WorkRoot}/target/worktree-A
  -> wrk myrepo {WorkRoot}/target
  -> new path; no would reuse
```

## Steps

1. Add clean linked sibling at `{WorkRoot}/target/worktree-A` (branch free of main-{date}).
2. Dirty it with an untracked file.
3. Run named create into existing `{WorkRoot}/target`.

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	parent := policyBTargetParent(req)
	sib := policyBAddSiblingUnderParent(t, req, parent, "worktree-A", "policy-b-dirty-sib")
	req.WtDir = sib
	// Dirty: untracked file (porcelain non-empty).
	writeFile(t, filepath.Join(sib, "dirty-untracked.txt"), "dirty\n")
	return nil
}
```
