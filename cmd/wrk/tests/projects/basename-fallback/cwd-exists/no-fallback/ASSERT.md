## Expected

- Non-zero exit code.
- Stderr mentions `not a git repository` (cwd path used, not saved project).
- Stdout is empty.
- No worktree created under `{WRK_HOME}/worktrees/`.

## Errors

- `./myrepo` exists in cwd but is not a git repository; fallback must not run.

## Exit Code

- Non-zero

```go
import (
	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
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

	wantPath := worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 0)
	assertFileNotExists(t, wantPath)
	assertWorktreeListNotContains(t, req.MainRepo, wantPath)
}
```