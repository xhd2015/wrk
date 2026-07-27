## Expected

- Exit code 0.
- Stdout is `…/external/mydep2-main-{WRK_DATE}` (new path for dep2).
- Prior `mydep1` external path still exists; not printed as stdout for this run.
- Exactly two external dirs (one per dep).
- Stderr has **no** reuse warning about reusing the mydep1 path.
- go.mod has replaces for both modules at their respective external paths.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantDep2 := bringExternalWorktreePath(req.ConsumerTop, "mydep2", "main", 0)
	priorDep1 := req.ExternalWtDir
	assertStdoutExactPath(t, resp.Stdout, wantDep2)

	if n := countExternalDirs(t, req.ConsumerTop); n != 2 {
		t.Fatalf("expected 2 external/ entries, got %d", n)
	}
	assertFileExists(t, priorDep1)
	assertFileExists(t, wantDep2)
	assertWorktreeListContains(t, req.DepPath, wantDep2)

	mod, err := readBringGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !bringHasReplaceForModule(mod, "example.com/dep1", priorDep1) {
		t.Fatalf("missing dep1 replace => %s: %+v", priorDep1, mod.Replace)
	}
	if !bringHasReplaceForModule(mod, "example.com/dep2", wantDep2) {
		t.Fatalf("missing dep2 replace => %s: %+v", wantDep2, mod.Replace)
	}

	// Must not claim to reuse the other dep's external path.
	assertNotContains(t, resp.Stderr, "reusing "+priorDep1)
	assertNotContains(t, resp.Stdout, priorDep1)
}
```
