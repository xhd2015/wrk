## Expected

- Exit code 0 (soft skip — unlike strict `--dep`).
- Stdout equals abs external path under `consumer/external/`.
- External worktree exists and is owned by the dep repo.
- Consumer `.gitignore` contains `/external`.
- Consumer still has no `go.mod` (no accidental module create).
- Stderr contains `SKIP local dep replacement` and `consumer has no Go modules`.
- Stderr does **not** use the "not a dependency" wording (distinct notice).

## Exit Code

- 0

```go
import (
	"path/filepath"
	"os"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
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
	assertWorktreeListNotContains(t, req.ConsumerTop, wantPath)

	ok, err := bringGitignoreContainsExternal(req.ConsumerTop)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !ok {
		t.Fatalf(".gitignore should contain /external")
	}

	if _, err := os.Stat(filepath.Join(req.ConsumerTop, "go.mod")); err == nil {
		t.Fatalf("consumer should still have no go.mod")
	}

	assertContains(t, resp.Stderr, "SKIP local dep replacement")
	assertContains(t, resp.Stderr, "consumer has no Go modules")
	assertNotContains(t, resp.Stderr, "not a dependency of any consumer module")
}
```
