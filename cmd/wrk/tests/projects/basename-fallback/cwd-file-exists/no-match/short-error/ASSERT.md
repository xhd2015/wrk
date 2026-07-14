## Expected

- Non-zero exit code.
- Stderr is a single line reporting cwd file path `exists and is a file`.
- Stderr does not list registered projects or emit a suggested command hint.
- Stdout is empty.
- No worktree created.

## Errors

- Cwd file blocks basename resolution and no saved project matches.

## Exit Code

- Non-zero

```go
import (
	"path/filepath"
	"strings"

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

	filePath := resolvePath(t, filepath.Join(req.RepoDir, "foo"))
	assert.Output(t, resp.Stderr, `<contains>
`+filePath+` exists and is a file
</contains>`)

	assertNotContains(t, resp.Stderr, "matches registered project")
	assertNotContains(t, resp.Stderr, "use `wrk")

	lines := strings.Split(strings.TrimSpace(resp.Stderr), "\n")
	if len(lines) != 1 {
		t.Fatalf("expected single stderr line, got %d:\n%s", len(lines), resp.Stderr)
	}

	wantPath := worktreePath(req.WrkHome, "foo", "main", wrkDate, 0)
	assertFileNotExists(t, wantPath)
}
```