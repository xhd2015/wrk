## Expected

- Exit code 0.
- Stdout: two external abs paths (mydep1 then mydep2).
- Both worktrees exist; `/external` gitignore present.
- Consumer `go.mod` is **byte-identical** to snapshot (no replaces).
- Stderr has no `SKIP local dep replacement` and no `mod tidy`.

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

	multiAssertBringGoModUnchanged(t, req, req.RepoDir)
	mod, err := readBringGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if bringHasAnyReplaceForModule(mod, multiBringDep1Module) {
		t.Fatalf("go.mod should have no replace for %s under --no-dep: %+v", multiBringDep1Module, mod.Replace)
	}
	if bringHasAnyReplaceForModule(mod, multiBringDep2Module) {
		t.Fatalf("go.mod should have no replace for %s under --no-dep: %+v", multiBringDep2Module, mod.Replace)
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
