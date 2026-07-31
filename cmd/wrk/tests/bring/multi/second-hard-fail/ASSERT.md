## Expected

- Non-zero exit code (hard error; fail-fast).
- First external worktree for mydep1 **exists** (earlier success kept).
- Exactly **one** external entry; second path not materialized.
- Stdout includes the first external abs path (partial success output allowed).
- Stderr indicates the second path problem (`does not exist` or equivalent).

## Exit Code

- Non-zero

```go
import (
	"strings"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit for second hard-fail, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}

	want1 := bringExternalWorktreePath(req.ConsumerTop, "mydep1", "main", 0)
	req.ExternalWtDir = want1
	assertFileExists(t, want1)
	assertGitFileIsWorktreeLink(t, want1)
	assertWorktreeListContains(t, req.DepPath, want1)

	if n := multiCountExternalDirs(t, req.ConsumerTop); n != 1 {
		t.Fatalf("expected exactly 1 external/ entry after fail-fast, got %d", n)
	}
	assertFileNotExists(t, bringExternalWorktreePath(req.ConsumerTop, "missing-dep-does-not-exist", "main", 0))

	if !strings.Contains(resp.Stdout, want1) {
		t.Fatalf("stdout should include first successful path %q; got %q", want1, resp.Stdout)
	}

	assertContains(t, resp.Stderr, "does not exist")
}
```
