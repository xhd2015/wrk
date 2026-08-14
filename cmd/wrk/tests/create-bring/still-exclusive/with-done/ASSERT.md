## Expected

- Non-zero exit.
- Stderr mentions mutual exclusion.
- No new worktree under `{WRK_HOME}/worktrees`.
- No `src/external/`.

## Errors

- `--done` and `--bring` cannot be combined.

## Exit Code

- Non-zero

```go
import (
	"path/filepath"
	"strings"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for --done --bring, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	se := strings.ToLower(resp.Stderr)
	if !strings.Contains(se, "mutually exclusive") && !strings.Contains(se, "exclusive") {
		t.Fatalf("stderr should mention mutual exclusion, got %q", resp.Stderr)
	}
	assertFileNotExists(t, createBringDefaultWT(req))
	assertFileNotExists(t, filepath.Join(req.MainRepo, "external"))
	if names := createBringListHomeWTs(t, req); len(names) != 0 {
		t.Fatalf("--done --bring must not create worktrees; found %v", names)
	}
}
```
