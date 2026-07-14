## Expected

- Exit code 0 (soft skip — unlike strict `--dep`).
- Stdout equals abs external path under `consumer/external/`.
- External worktree exists and is owned by the dep repo.
- Consumer `.gitignore` contains `/external`.
- Consumer `go.mod` has **no** replace for `example.com/dep`.
- Stderr contains `SKIP local dep replacement` and `is not a go module`.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
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

	ok, err := bringGitignoreContainsExternal(req.ConsumerTop)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !ok {
		t.Fatalf(".gitignore should contain /external")
	}

	mod, err := readBringGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if bringHasAnyReplaceForModule(mod, bringDepModulePath) {
		t.Fatalf("go.mod should not replace %s after SKIP: %+v", bringDepModulePath, mod.Replace)
	}

	assertContains(t, resp.Stderr, "SKIP local dep replacement")
	assertContains(t, resp.Stderr, "is not a go module")
}
```
