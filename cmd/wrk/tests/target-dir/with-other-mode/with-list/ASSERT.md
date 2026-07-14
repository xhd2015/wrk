## Expected

- Non-zero exit code.
- Stderr is exactly `wrk: unexpected arguments` (target-dir is create-only).
- Stdout is empty.
- No worktree is created and `--list` does not run (no worktree-list output on stdout).

## Errors

- `<target-dir>` combined with `--list`.

## Exit Code

- Non-zero

```go
import (
	"strings"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit code, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if strings.TrimSpace(resp.Stdout) != "" {
		t.Fatalf("stdout should be empty, got %q", resp.Stdout)
	}
	assert.Output(t, resp.Stderr, `<contains>
wrk: unexpected arguments
</contains>`)

	// no worktree created
	assertFileNotExists(t, req.SpawnDir)
	assertFileNotExists(t, worktreePath(req.WrkHome, "myrepo", "main", wrkDate, 0))
}
```
