## Expected

- Exit code 0 (default auto-skip; non-TTY no longer hard-refuses).
- Stdout (trimmed) equals the existing linked worktree path.
- No new worktree under `{WorkRoot}/target/`.
- Prior worktree unchanged.
- No `Proceed?` / interactive skip prompt required (silent auto-skip OK).

## Exit Code

- 0

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 on non-TTY named bring auto-skip; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}

	wantPath := req.WtDir
	assertStdoutExactPath(t, resp.Stdout, wantPath)

	assertFileExists(t, wantPath)
	assertFileNotExists(t, filepath.Join(req.WorkRoot, "target", "myrepo-main-"+wrkDate))
	assertFileNotExists(t, filepath.Join(req.WorkRoot, "target", "myrepo-main-"+wrkDate+"-1"))
}
```
