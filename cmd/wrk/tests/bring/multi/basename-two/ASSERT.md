## Expected

- Exit code 0.
- Stdout: two lines — `…/external/mydep1-main-{date}` then `…/external/mydep2-main-{date}`.
- Both worktrees owned by their saved dep mains; replaces for both modules.
- `/external` gitignore present; no SKIP.

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

	assertFileExists(t, want1)
	assertFileExists(t, want2)
	assertWorktreeListContains(t, req.DepPath, want1)
	assertWorktreeListContains(t, req.SecondRepo, want2)

	mod, err := readBringGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !bringHasReplaceForModule(mod, multiBringDep1Module, want1) {
		t.Fatalf("missing replace %s => %s: %+v", multiBringDep1Module, want1, mod.Replace)
	}
	if !bringHasReplaceForModule(mod, multiBringDep2Module, want2) {
		t.Fatalf("missing replace %s => %s: %+v", multiBringDep2Module, want2, mod.Replace)
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
