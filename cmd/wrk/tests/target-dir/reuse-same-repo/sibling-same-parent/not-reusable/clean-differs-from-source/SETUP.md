# Scenario

**Feature**: clean same-parent sibling whose HEAD differs from source is not reusable → create, no banner

```
# sibling worktree-A clean but has extra commit (ahead of source HEAD)
# wrk myrepo target -> create; no would-reuse / skip-creating
myrepo + clean-differs sibling {WorkRoot}/target/worktree-A
  -> wrk myrepo {WorkRoot}/target
  -> new path; no Policy B banner
```

## Steps

1. Add linked sibling at `{WorkRoot}/target/worktree-A`.
2. Commit ahead on the sibling so HEAD differs from source main.
3. Ensure sibling porcelain is clean after the commit.
4. Run named create into existing target.

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	parent := policyBTargetParent(req)
	sib := policyBAddSiblingUnderParent(t, req, parent, "worktree-A", "policy-b-differs-sib")
	req.WtDir = sib
	srcHEAD := revParseHEAD(t, req.TargetDir)
	commitAheadOnWorktree(t, sib, "ahead.txt", "differs from source\n")
	sibHEAD := revParseHEAD(t, sib)
	if sibHEAD == srcHEAD {
		t.Fatalf("expected sibling HEAD to differ from source; both %s", srcHEAD)
	}
	req.MainRepo = srcHEAD // stash source HEAD for assert (field reuse)
	// porcelain clean after commit
	status := gitOutputIsolated(t, sib, "status", "--porcelain")
	if strings.TrimSpace(status) != "" {
		t.Fatalf("sibling should be clean after commit; porcelain=%q", status)
	}
	return nil
}
```
