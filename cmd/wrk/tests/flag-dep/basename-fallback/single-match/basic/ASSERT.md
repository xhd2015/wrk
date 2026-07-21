## Expected

- Exit code 0.
- Stdout (trimmed) equals `{consumerTop}/external/mydep-main-{WRK_DATE}`.
- External path exists as a linked git worktree owned by the **saved** dep repo.
- Consumer `go.mod` has `replace example.com/dep => <external path>`.
- Consumer `.gitignore` contains `/external`.

## Side Effects

- External worktree created from saved project path resolved via basename fallback, not from a local `./mydep` path.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantPath := externalWorktreePath(req.ConsumerTop, "mydep", "main", 0)
	req.ExternalWtDir = wantPath
	assertStdoutExactPath(t, resp.Stdout, wantPath)

	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)
	assertWorktreeListContains(t, req.DepPath, wantPath)
	assertWorktreeListNotContains(t, req.ConsumerTop, wantPath)

	mod, err := readGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if !hasReplaceForModule(mod, depModulePath, wantPath) {
		t.Fatalf("go.mod missing replace %s => %s: %+v", depModulePath, wantPath, mod.Replace)
	}

	ok, err := gitignoreContainsExternal(req.ConsumerTop)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !ok {
		t.Fatalf(".gitignore should contain /external")
	}

	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
}
```