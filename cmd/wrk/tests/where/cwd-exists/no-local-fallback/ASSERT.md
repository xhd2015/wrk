## Expected Output

Saved project absolute path only (not the cwd `./spl` directory path).

## Expected

- Exit code 0.
- Stdout equals the **saved** project's absolute path plus trailing `\n`.
- Stdout does not contain the cwd local `./spl` path.
- Stderr is empty.
- No worktree created.

## Side Effects

- Read-only lookup from `projects.json`; cwd directory is not consulted.

## Exit Code

- 0

```go
import (
	"path/filepath"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	savedPath := resolvePath(t, req.MainRepo)
	localPath := resolvePath(t, filepath.Join(req.RepoDir, whereBasename))
	assertStdoutExactPath(t, resp.Stdout, savedPath)
	assertNotContains(t, resp.Stdout, localPath)

	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}

	wantPath := worktreePath(req.WrkHome, whereBasename, "main", wrkDate, 0)
	assertFileNotExists(t, wantPath)
}
```