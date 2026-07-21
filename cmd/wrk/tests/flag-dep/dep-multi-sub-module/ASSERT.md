## Expected

- Exit code 0.
- Stdout (trimmed) equals `{consumerTop}/external/mydep-main-{WRK_DATE}`.
- External path exists as a linked git worktree.
- Consumer `go.mod` has `replace example.com/dep/b => <external>/b`.
- Consumer `go.mod` has NO replace for `example.com/dep/a`.

## Exit Code

- 0

```go
import "path/filepath"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
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

	// replace must target b/ sub-directory inside external worktree
	wantReplacePath := filepath.Join(wantPath, "b")
	mod, err := readGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !hasReplaceForModule(mod, depModulePathB, wantReplacePath) {
		t.Fatalf("go.mod missing replace %s => %s: %+v", depModulePathB, wantReplacePath, mod.Replace)
	}
	// must NOT have a replace for the unmatched sub-module
	if hasReplaceForModule(mod, depModulePathA, filepath.Join(wantPath, "a")) {
		t.Fatalf("go.mod should NOT have replace for %s: %+v", depModulePathA, mod.Replace)
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