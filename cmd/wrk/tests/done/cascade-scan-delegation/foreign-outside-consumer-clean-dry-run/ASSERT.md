## Expected

- Exit code 0 (dry-run preflight does not treat foreign dirty as cascade target).
- Zero mutations: consumer worktree still present and linked; branch still exists.
- Foreign dirty tree still present.
- Stderr/stdout do **not** mention foreign path or `other/external/agent-pro`.
- Must not print `would: cascade` for the foreign tree (zero in-consumer cascade).
- Primary dry-run plan may appear (`merge --ff-only`, `worktree remove`, `branch -D`);
  do not require phase headers (zero cascade targets).

## Exit Code

- 0

```go
import (
	"path/filepath"
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 (foreign dirty must not block --done --dry-run), got %d stdout=%q stderr=%q",
			resp.ExitCode, resp.Stdout, resp.Stderr)
	}

	assertNoForeignPathInOutput(t, req, resp)

	combined := resp.Stdout + "\n" + resp.Stderr
	// Foreign must never be planned as a cascade target.
	if strings.Contains(combined, "would: cascade") &&
		(strings.Contains(combined, req.SecondRepo) ||
			strings.Contains(combined, filepath.Base(req.SecondRepo))) {
		t.Fatalf("dry-run must not cascade-plan foreign path; combined:\n%s", combined)
	}

	// Zero mutations on consumer.
	assertFileExists(t, req.WtDir)
	assertGitFileIsWorktreeLink(t, req.WtDir)
	assertWorktreeListContains(t, req.MainRepo, req.WtDir)
	assertBranchExists(t, req.MainRepo, req.WtBranch)

	// Foreign untouched.
	assertFileExists(t, req.SecondRepo)
	assertFileExists(t, filepath.Join(req.SecondRepo, "dirty-foreign"))

	// Soft: primary dry-run vocabulary when product prints a plan (not required
	// if zero-cascade path omits plan body — exit 0 + no foreign leak is enough).
	_ = strings.Contains(resp.Stdout, "merge --ff-only")
}
```
