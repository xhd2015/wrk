## Expected

- Exit code 0.
- Stdout is the lex-smallest external path (`…/mydep-main-{date}`, not `…-1`).
- Still exactly **two** entries under `external/` (no third collision dir).
- Stderr multi-reuse warning mentions count `2` and `reusing` the smallest path.
- Stderr also lists the other path (`also present:` + second abs path).
- Replace still points at the reused (smallest) path.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantPath := req.ExternalWtDir
	other := req.ExternalWtDir2
	assertStdoutExactPath(t, resp.Stdout, wantPath)

	if n := countExternalDirs(t, req.ConsumerTop); n != 2 {
		t.Fatalf("expected exactly 2 external/ entries, got %d", n)
	}
	assertFileNotExists(t, bringExternalWorktreePath(req.ConsumerTop, "mydep", "main", 2))

	assertFileExists(t, wantPath)
	assertFileExists(t, other)
	assertWorktreeListContains(t, req.DepPath, wantPath)
	assertWorktreeListContains(t, req.DepPath, other)

	mod, err := readBringGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !bringHasReplaceForModule(mod, bringDepModulePath, wantPath) {
		t.Fatalf("go.mod replace should still target reused path %s: %+v", wantPath, mod.Replace)
	}

	assertContains(t, resp.Stderr, "already has")
	assertContains(t, resp.Stderr, "2")
	assertContains(t, resp.Stderr, "worktrees under external/")
	assertContains(t, resp.Stderr, "reusing")
	assertContains(t, resp.Stderr, wantPath)
	assertContains(t, resp.Stderr, "also present")
	assertContains(t, resp.Stderr, other)
}
```
