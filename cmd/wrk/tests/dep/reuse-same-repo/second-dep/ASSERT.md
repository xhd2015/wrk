## Expected

- Exit code 0.
- Stdout equals the first external path.
- Exactly one `external/` entry (no `…-1`).
- Replace `example.com/dep => <first path>` still present.
- Branch still `main-{WRK_DATE}` on dep (no second branch from reuse path).
- Stderr contains reuse warning with basename and abs path.

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

	wantPath := req.ExternalWtDir
	assertStdoutExactPath(t, resp.Stdout, wantPath)

	if n := countDepExternalDirs(t, req.ConsumerTop); n != 1 {
		t.Fatalf("expected exactly 1 external/ entry, got %d", n)
	}
	assertFileNotExists(t, externalWorktreePath(req.ConsumerTop, "mydep", "main", 1))

	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)
	assertWorktreeListContains(t, req.DepPath, wantPath)

	wantBranch := branchName("main", wrkDate, 0)
	assertBranchExists(t, req.DepPath, wantBranch)
	assertBranchNotExists(t, req.DepPath, branchName("main", wrkDate, 1))

	mod, err := readGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !hasReplaceForModule(mod, depModulePath, wantPath) {
		t.Fatalf("go.mod missing replace %s => %s: %+v", depModulePath, wantPath, mod.Replace)
	}

	ok, err := gitignoreContainsExternal(req.ConsumerTop)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !ok {
		t.Fatalf(".gitignore should contain /external")
	}

	assertContains(t, resp.Stderr, "already exists under external/")
	assertContains(t, resp.Stderr, "reusing")
	assertContains(t, resp.Stderr, wantPath)
	assertContains(t, resp.Stderr, "mydep")
}
```
