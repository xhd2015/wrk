## Expected

- Exit code 0.
- Stdout matches `git -C myrepo worktree list` (same as from repo root).
- Stdout matches `git -C subpath worktree list`.
- Stderr is empty.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}

	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	wantRoot := gitWorktreeListIsolated(t, mainRepo)
	wantSubpath := gitWorktreeListIsolated(t, req.RepoDir)

	if wantRoot != wantSubpath {
		t.Fatalf("git worktree list should match from root and subpath:\nroot:\n%q\nsubpath:\n%q", wantRoot, wantSubpath)
	}

	if resp.Stdout != wantRoot {
		t.Fatalf("stdout mismatch:\nwant:\n%q\ngot:\n%q", wantRoot, resp.Stdout)
	}

	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
}
```