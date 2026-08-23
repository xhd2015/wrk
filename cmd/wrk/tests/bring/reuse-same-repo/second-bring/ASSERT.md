## Expected

- Exit code 0.
- Stdout (exact path + `\n`) equals the **first** external path (`…/external/mydep`).
- Only **one** entry under `{consumerTop}/external/` (no `…-1` collision dir).
- First external path still exists as a linked worktree of the dep main.
- Branch remains `main-{WRK_DATE}` on the dep (no new `main-{date}-1` from the second bring).
- Consumer `go.mod` still has `replace example.com/dep => <first external path>`.
- Stderr contains reuse warning with basename `mydep` and the reused abs path.
- Stderr does **not** require SKIP notices for this happy path.

## Side Effects

- No second `git worktree add` for the same dep under `external/`.
- `/external` gitignore remains present.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantPath := req.ExternalWtDir
	assertStdoutExactPath(t, resp.Stdout, wantPath)

	if n := countExternalDirs(t, req.ConsumerTop); n != 1 {
		t.Fatalf("expected exactly 1 external/ entry, got %d", n)
	}
	collided := bringExternalWorktreePath(req.ConsumerTop, "mydep", "main", 1)
	assertFileNotExists(t, collided)

	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)
	assertWorktreeListContains(t, req.DepPath, wantPath)
	assertWorktreeListNotContains(t, req.ConsumerTop, wantPath)

	wantBranch := branchName("main", wrkDate, 0)
	assertBranchExists(t, req.DepPath, wantBranch)
	assertBranchNotExists(t, req.DepPath, branchName("main", wrkDate, 1))
	assertBranchCheckedOutInWorktree(t, wantPath, wantBranch)

	mod, err := readBringGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !bringHasReplaceForModule(mod, bringDepModulePath, wantPath) {
		t.Fatalf("go.mod missing replace %s => %s: %+v", bringDepModulePath, wantPath, mod.Replace)
	}

	assertContains(t, resp.Stderr, "already exists under external/")
	assertContains(t, resp.Stderr, "reusing")
	assertContains(t, resp.Stderr, wantPath)
	assertContains(t, resp.Stderr, "mydep")
	assertNotContains(t, resp.Stderr, "SKIP local dep replacement")
}
```
