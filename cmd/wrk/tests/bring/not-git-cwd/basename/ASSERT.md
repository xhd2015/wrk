## Expected

- Exit code 0.
- Stdout equals `{plain}/external/mydep-main-{WRK_DATE}` (basename resolves via projects).
- External worktree exists and is owned by the **saved** dep repo.
- **No** `.gitignore` under the plain cwd.
- Stderr contains `SKIP local dep replacement` and `is not a git repository`.

## Side Effects

- External worktree created from the registered project path, not from a local `./mydep` under the plain cwd.

## Exit Code

- 0

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantPath := bringExternalWorktreePath(req.ConsumerTop, "mydep", "main", 0)
	req.ExternalWtDir = wantPath
	assertStdoutExactPath(t, resp.Stdout, wantPath)

	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)
	assertWorktreeListContains(t, req.DepPath, wantPath)

	assertFileNotExists(t, filepath.Join(req.ConsumerTop, ".gitignore"))

	assertContains(t, resp.Stderr, "SKIP local dep replacement")
	assertContains(t, resp.Stderr, "is not a git repository")
	assertContains(t, resp.Stderr, req.RepoDir)
}
```
