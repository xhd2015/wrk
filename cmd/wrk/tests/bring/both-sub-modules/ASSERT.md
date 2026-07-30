## Expected

- Exit code 0.
- Stdout (trimmed) equals `{consumerTop}/external/mydep-main-{WRK_DATE}`.
- External path exists as a linked git worktree.
- Consumer `go-pkgs/go.mod` has `replace example.com/dep/sub => <external>/sub`.
- Consumer `.gitignore` contains `/external`.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/session"
)

import "path/filepath"

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
	assertWorktreeListNotContains(t, req.ConsumerTop, wantPath)

	// the sub-module's go.mod must be checked out inside the external worktree
	assertFileExists(t, filepath.Join(wantPath, "sub", "go.mod"))

	// replace must target the sub-module directory under external
	wantReplacePath := filepath.Join(wantPath, "sub")
	mod, err := readBringGoMod(req.ConsumerModDir)
	if err != nil {
		t.Fatalf("read go-pkgs/go.mod: %v", err)
	}
	if !bringHasReplaceForModule(mod, depSubModulePath, wantReplacePath) {
		t.Fatalf("go-pkgs/go.mod missing replace %s => %s: %+v", depSubModulePath, wantReplacePath, mod.Replace)
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