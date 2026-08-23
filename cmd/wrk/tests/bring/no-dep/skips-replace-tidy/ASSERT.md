## Expected

- Exit code 0.
- Stdout (trimmed) equals `{consumerTop}/external/mydep`.
- External path exists as a linked git worktree owned by the **dep** main repo.
- Branch in the dep repo is `main-{WRK_DATE}`.
- Consumer `go.mod` is **byte-identical** to pre-run snapshot (no new replace).
- Consumer `.gitignore` contains `/external`.
- Stderr does **not** contain `SKIP local dep replacement` (analyse skipped).
- Stderr does **not** contain `mod tidy`.

## Exit Code

- 0

```go
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

	wantBranch := branchName("main", wrkDate, 0)
	assertBranchExists(t, req.DepPath, wantBranch)
	assertBranchCheckedOutInWorktree(t, wantPath, wantBranch)

	assertBringGoModUnchanged(t, req, req.RepoDir)
	mod, err := readBringGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if bringHasAnyReplaceForModule(mod, bringDepModulePath) {
		t.Fatalf("go.mod should have no replace for %s under --no-dep: %+v", bringDepModulePath, mod.Replace)
	}

	ok, err := bringGitignoreContainsExternal(req.ConsumerTop)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !ok {
		t.Fatalf(".gitignore should contain /external")
	}

	assertNotContains(t, resp.Stderr, "SKIP local dep replacement")
	assertNotContains(t, resp.Stderr, "mod tidy")
}
```
