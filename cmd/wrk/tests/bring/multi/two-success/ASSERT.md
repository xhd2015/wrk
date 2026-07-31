## Expected

- Exit code 0.
- Stdout is exactly two lines: external path for mydep1 then mydep2 (left→right bring order), trailing `\n`.
- Both external paths exist as linked worktrees owned by their respective dep mains.
- Branches are `main-{WRK_DATE}` on each dep (no basename prefix).
- Consumer `go.mod` has replaces for both modules at their external paths.
- Consumer `.gitignore` contains `/external`.
- Stderr does **not** contain `SKIP local dep replacement`.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	want1 := bringExternalWorktreePath(req.ConsumerTop, "mydep1", "main", 0)
	want2 := bringExternalWorktreePath(req.ConsumerTop, "mydep2", "main", 0)
	req.ExternalWtDir = want1
	req.ExternalWtDir2 = want2
	assertStdoutTwoPathsExact(t, resp.Stdout, want1, want2)

	if n := multiCountExternalDirs(t, req.ConsumerTop); n != 2 {
		t.Fatalf("expected 2 external/ entries, got %d", n)
	}

	assertFileExists(t, want1)
	assertFileExists(t, want2)
	assertGitFileIsWorktreeLink(t, want1)
	assertGitFileIsWorktreeLink(t, want2)
	assertWorktreeListContains(t, req.DepPath, want1)
	assertWorktreeListContains(t, req.SecondRepo, want2)
	assertWorktreeListNotContains(t, req.ConsumerTop, want1)
	assertWorktreeListNotContains(t, req.ConsumerTop, want2)

	wantBranch := branchName("main", wrkDate, 0)
	assertBranchExists(t, req.DepPath, wantBranch)
	assertBranchExists(t, req.SecondRepo, wantBranch)
	assertBranchCheckedOutInWorktree(t, want1, wantBranch)
	assertBranchCheckedOutInWorktree(t, want2, wantBranch)

	mod, err := readBringGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !bringHasReplaceForModule(mod, multiBringDep1Module, want1) {
		t.Fatalf("go.mod missing replace %s => %s: %+v", multiBringDep1Module, want1, mod.Replace)
	}
	if !bringHasReplaceForModule(mod, multiBringDep2Module, want2) {
		t.Fatalf("go.mod missing replace %s => %s: %+v", multiBringDep2Module, want2, mod.Replace)
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
