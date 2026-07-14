## Expected

- Exit code 0.
- Stdout (trimmed) equals `{consumerTop}/external/mydep-main-{WRK_DATE}`.
- External path exists as a linked git worktree owned by the **dep** main repo.
- Branch in the dep repo is `main-{WRK_DATE}` (**no** dep basename prefix on the branch).
- Consumer `go.mod` has `replace example.com/dep => <external path>`.
- Consumer `.gitignore` contains `/external`.
- Stderr does **not** contain `SKIP local dep replacement`.

## Exit Code

- 0

```go
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
	// The external dep worktree is owned by the DEP repo, not the consumer.
	assertWorktreeListContains(t, req.DepPath, wantPath)
	assertWorktreeListNotContains(t, req.ConsumerTop, wantPath)

	wantBranch := branchName("main", wrkDate, 0)
	assertBranchExists(t, req.DepPath, wantBranch)
	assertBranchNotExists(t, req.DepPath, "mydep-"+wantBranch)
	assertBranchCheckedOutInWorktree(t, wantPath, wantBranch)

	mod, err := readBringGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !bringHasReplaceForModule(mod, bringDepModulePath, wantPath) {
		t.Fatalf("go.mod missing replace %s => %s: %+v", bringDepModulePath, wantPath, mod.Replace)
	}

	ok, err := bringGitignoreContainsExternal(req.ConsumerTop)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !ok {
		t.Fatalf(".gitignore should contain /external")
	}

	assertNotContains(t, resp.Stderr, "SKIP local dep replacement")
}
```
