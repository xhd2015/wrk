## Expected

- Non-zero exit code.
- Stderr mentions `not a git repository` (local cwd path used, not saved project).
- Stdout is empty.
- No external worktree created under `{consumerTop}/external/`.

## Errors

- `./mydep` exists in consumer cwd but is not a git repository; fallback must not run.

## Exit Code

- Non-zero

```go
import (
	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	assert.Output(t, resp.Stderr, `<contains>
not a git repository
</contains>`)

	wantPath := bringExternalWorktreePath(req.ConsumerTop, "mydep", "main", 0)
	assertFileNotExists(t, wantPath)
	assertWorktreeListNotContains(t, req.DepPath, wantPath)
}
```