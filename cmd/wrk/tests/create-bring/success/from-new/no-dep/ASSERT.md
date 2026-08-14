## Expected

- Exit code 0.
- New WT and external worktree exist.
- New WT `go.mod` has **no** replace for dep1.
- Source `src/go.mod` unchanged.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wt := createBringDefaultWT(req)
	ext1 := createBringExternalPath(wt, "mydep1")
	assertFileExists(t, wt)
	assertFileExists(t, ext1)
	if !createBringStdoutHasLine(resp.Stdout, wt) {
		t.Fatalf("stdout should include create path %q; got %q", wt, resp.Stdout)
	}
	if !createBringStdoutHasLine(resp.Stdout, ext1) {
		t.Fatalf("stdout should include external %q; got %q", ext1, resp.Stdout)
	}

	createBringAssertGoModUnchanged(t, req, req.MainRepo)
	mod, err := readCreateBringGoMod(wt)
	if err != nil {
		t.Fatalf("read new WT go.mod: %v", err)
	}
	if createBringHasAnyReplace(mod, createBringDep1Module) {
		t.Fatalf("new WT go.mod should have no replace for %s under --no-dep: %+v", createBringDep1Module, mod.Replace)
	}
	ok, err := createBringGitignoreHasExternal(wt)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !ok {
		t.Fatalf("new WT .gitignore should contain /external")
	}
}
```
