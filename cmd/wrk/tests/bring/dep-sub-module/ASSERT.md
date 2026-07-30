## Expected

- Exit code 0.
- Stdout (trimmed) equals `{consumerTop}/external/mydep-main-{WRK_DATE}` (the dep repo worktree).
- External path exists as a linked git worktree.
- The sub-module's `go.mod` exists inside the external worktree at `<external>/sub/go.mod`.
- Consumer `go.mod` has `replace example.com/dep/sub => <external>/sub` (the directory
  holding the sub-module's `go.mod`, not the repo root).
- Consumer `.gitignore` contains `/external`.

## Exit Code

- 0

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
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
	// The external dep worktree is owned by the DEP repo, not the consumer.
	assertWorktreeListContains(t, req.DepPath, wantPath)
	assertWorktreeListNotContains(t, req.ConsumerTop, wantPath)

	// the sub-module's go.mod must be checked out inside the external worktree
	assertFileExists(t, filepath.Join(wantPath, "sub", "go.mod"))

	mod, err := readBringGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	// replace must target the sub-module directory (where go.mod lives), not the repo root
	wantReplacePath := filepath.Join(wantPath, "sub")
	if !bringHasReplaceForModule(mod, subModulePath, wantReplacePath) {
		t.Fatalf("go.mod missing replace %s => %s: %+v", subModulePath, wantReplacePath, mod.Replace)
	}

	ok, err := bringGitignoreContainsExternal(req.ConsumerTop)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !ok {
		t.Fatalf(".gitignore should contain /external")
	}
}
```
