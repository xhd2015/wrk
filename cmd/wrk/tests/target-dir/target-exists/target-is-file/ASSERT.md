## Expected

- Non-zero exit code.
- Stderr indicates `<target-dir>` is not a directory / cannot spawn there.
- Stdout is empty.
- No worktree is created (the file at `<target-dir>` is left as-is; nothing under `{WRK_HOME}`).

## Errors

- `<target-dir>` exists but is not a directory.

## Exit Code

- Non-zero

```go
import (
	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit code, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	assert.Output(t, resp.Stderr, `<contains>
not a directory
</contains>`)

	// the file is untouched; no worktree created anywhere
	assertFileExists(t, req.SpawnDir)
	assertFileNotExists(t, worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 0))
}
```
