## Expected

- Exit code 0.
- Stdout abs external path; worktree owned by dep.
- Branch `main-{WRK_DATE}`; gitignore `/external`.
- Consumer go.mod byte-identical (no replace for `example.com/dep`).

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantPath := externalWorktreePath(req.ConsumerTop, "mydep", "main", 0)
	req.ExternalWtDir = wantPath
	assertStdoutExactPath(t, resp.Stdout, wantPath)

	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)
	assertWorktreeListContains(t, req.DepPath, wantPath)
	assertWorktreeListNotContains(t, req.ConsumerTop, wantPath)

	wantBranch := branchName("main", wrkDate, 0)
	assertBranchExists(t, req.DepPath, wantBranch)
	assertBranchCheckedOutInWorktree(t, wantPath, wantBranch)

	assertDepGoModUnchanged(t, req, req.RepoDir)
	mod, err := readGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if hasReplaceForModule(mod, depModulePath, wantPath) {
		t.Fatalf("go.mod should not replace %s under --no-dep: %+v", depModulePath, mod.Replace)
	}
	// Also reject any replace for the module path.
	for _, repl := range mod.Replace {
		if repl.Old.Path == depModulePath {
			t.Fatalf("go.mod should have no replace for %s: %+v", depModulePath, mod.Replace)
		}
	}

	ok, err := gitignoreContainsExternal(req.ConsumerTop)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !ok {
		t.Fatalf(".gitignore should contain /external")
	}
}
```
