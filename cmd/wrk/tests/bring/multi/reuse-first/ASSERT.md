## Expected

- Exit code 0.
- Stdout: first line = reused mydep1 external path; second line = new mydep2 path.
- Exactly **two** external entries (no second mydep1 `-1` dir).
- Stderr contains reuse wording for mydep1 (`already exists under external/` + `reusing` + path).
- go.mod has replaces for both modules at their external paths.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	want1 := req.ExternalWtDir
	want2 := bringExternalWorktreePath(req.ConsumerTop, "mydep2", "main", 0)
	req.ExternalWtDir2 = want2
	assertStdoutTwoPathsExact(t, resp.Stdout, want1, want2)

	if n := multiCountExternalDirs(t, req.ConsumerTop); n != 2 {
		t.Fatalf("expected 2 external/ entries, got %d", n)
	}
	collided := bringExternalWorktreePath(req.ConsumerTop, "mydep1", "main", 1)
	assertFileNotExists(t, collided)
	assertFileExists(t, want1)
	assertFileExists(t, want2)
	assertWorktreeListContains(t, req.DepPath, want1)
	assertWorktreeListContains(t, req.SecondRepo, want2)

	mod, err := readBringGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !bringHasReplaceForModule(mod, multiBringDep1Module, want1) {
		t.Fatalf("missing dep1 replace => %s: %+v", want1, mod.Replace)
	}
	if !bringHasReplaceForModule(mod, multiBringDep2Module, want2) {
		t.Fatalf("missing dep2 replace => %s: %+v", want2, mod.Replace)
	}

	assertContains(t, resp.Stderr, "already exists under external/")
	assertContains(t, resp.Stderr, "reusing")
	assertContains(t, resp.Stderr, want1)
	assertContains(t, resp.Stderr, "mydep1")
}
```
