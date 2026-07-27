## Expected

- Exit code 0.
- Stdout (trimmed) equals `{consumerTop}/external/mydep-main-{WRK_DATE}`.
- External path exists as a linked git worktree.
- Branch in the dep repo is `main-{WRK_DATE}` (**no** dep basename prefix on the branch; P2).
- Consumer `go.mod` has `replace example.com/dep => <external path>`.
- Consumer `.gitignore` contains `/external`.

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
	// The external dep worktree is owned by the DEP repo, not the consumer.
	assertWorktreeListContains(t, req.DepPath, wantPath)
	assertWorktreeListNotContains(t, req.ConsumerTop, wantPath)

	// P2: branch is {token}-{date}, not {depBasename}-{token}-{date}.
	wantBranch := branchName("main", wrkDate, 0)
	assertBranchExists(t, req.DepPath, wantBranch)
	assertBranchNotExists(t, req.DepPath, "mydep-"+wantBranch)
	assertBranchCheckedOutInWorktree(t, wantPath, wantBranch)

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
}
```
