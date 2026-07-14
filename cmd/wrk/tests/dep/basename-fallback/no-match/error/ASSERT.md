## Expected

- Non-zero exit code.
- Stderr reports `does not exist` for the cwd-resolved candidate path.
- Stdout is empty.
- No external worktree created.

## Errors

- Basename missing from consumer cwd and no saved project match.

## Exit Code

- Non-zero

```go
import (
	"path/filepath"

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
does not exist
</contains>`)

	candidate := resolvePath(t, filepath.Join(req.RepoDir, "mydep"))
	assertContains(t, resp.Stderr, candidate)

	wantPath := externalWorktreePath(req.ConsumerTop, "mydep", "main", 0)
	assertFileNotExists(t, wantPath)
}
```