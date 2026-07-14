## Expected

- Non-zero exit code.
- Stderr explains that the (parent) path does not exist (mentions `does not exist`).
- Stdout is empty.
- No worktree is created anywhere (neither at `<target-dir>` nor under `{WRK_HOME}`).

## Errors

- `<target-dir>` parent directory does not exist.

## Exit Code

- Non-zero

```go
import (
	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit code, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if resp.Stdout != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	assert.Output(t, resp.Stderr, `<contains>
does not exist
</contains>`)

	// no worktree created
	assertFileNotExists(t, req.SpawnDir)
	assertFileNotExists(t, worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 0))
}
```
