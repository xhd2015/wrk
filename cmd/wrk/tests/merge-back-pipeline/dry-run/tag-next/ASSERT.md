## Expected

- Exit code 0.
- No confirm prompt / non-TTY confirm errors.
- Primary dry-run: `merge --ff-only <WtBranch>`; **no** `worktree remove` / `branch -D`.
- Tag-next dry-run block present: root-bump plan + `1 tag planned`.
- Blank line between primary plan region and tag plan (`1 tag planned` after merge plan).
- Zero mutations: source wt kept; main HEAD unchanged; no `v0.0.2`; no `feature-work` on main.
- No apply lines: `tag created` / `tagged `.

## Side Effects

- None (plan only); worktree remains.

## Exit Code

- 0

```go
import (
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertNotContains(t, resp.Stdout, "Proceed?")
	assertNotContains(t, resp.Stderr, "stdin is not a terminal")
	assertNotContains(t, resp.Stderr, "cannot prompt")

	assertPrimaryMergeBackKeepDryRunPlanned(t, resp.Stdout, req.WtBranch)

	tagBlock := strings.TrimSpace(tagNextRootBumpPlanStdoutMB())
	if !strings.Contains(resp.Stdout, tagBlock) {
		t.Fatalf("missing tag-next dry-run plan block %q\nstdout:\n%s", tagBlock, resp.Stdout)
	}
	idxMerge := strings.Index(resp.Stdout, "merge --ff-only "+req.WtBranch)
	idxTag := strings.Index(resp.Stdout, "1 tag planned")
	if idxMerge < 0 || idxTag < 0 || idxMerge > idxTag {
		t.Fatalf("want primary plan before tag plan; merge=%d tag=%d\n%s", idxMerge, idxTag, resp.Stdout)
	}

	assertNotContains(t, resp.Stdout, "tag created")
	assertNotContains(t, resp.Stdout, "tagged ")
	assertNotContains(t, resp.Stdout, "worktree removed:")

	assertMergeBackDryRunZeroMutations(t, req)
}
```
