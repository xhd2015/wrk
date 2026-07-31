## Expected

- Exit code 0 (soft SKIP continues; overall success).
- Stdout: two lines — external path for mydep1 then mydep2.
- Both external worktrees exist (worktree always materializes on soft SKIP).
- `go.mod` has replace for dep1 only; **no** replace for `example.com/dep2`.
- Stderr contains `SKIP local dep replacement` and `not a dependency of any consumer module`.
- `/external` gitignore present.

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
	assertWorktreeListContains(t, req.DepPath, want1)
	assertWorktreeListContains(t, req.SecondRepo, want2)

	mod, err := readBringGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !bringHasReplaceForModule(mod, multiBringDep1Module, want1) {
		t.Fatalf("go.mod missing replace %s => %s: %+v", multiBringDep1Module, want1, mod.Replace)
	}
	if bringHasAnyReplaceForModule(mod, multiBringDep2Module) {
		t.Fatalf("go.mod should not replace %s after soft SKIP: %+v", multiBringDep2Module, mod.Replace)
	}

	ok, err := bringGitignoreContainsExternal(req.ConsumerTop)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !ok {
		t.Fatalf(".gitignore should contain /external")
	}

	assertContains(t, resp.Stderr, "SKIP local dep replacement")
	assertContains(t, resp.Stderr, "not a dependency of any consumer module")
}
```
