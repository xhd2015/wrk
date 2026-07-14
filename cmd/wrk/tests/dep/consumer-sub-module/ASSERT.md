## Expected

- Exit code 0.
- Stdout (trimmed) equals `{consumerTop}/external/mydep-main-{WRK_DATE}`.
- External path exists as a linked git worktree.
- Consumer sub-module `go-pkgs/go.mod` has `replace example.com/dep => <external path>`.
- Consumer repo root has NO `go.mod` but the dep operation still succeeds via scanning.
- Consumer `.gitignore` contains `/external`.

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
	// The external dep worktree is owned by the DEP repo, not the consumer.
	assertWorktreeListContains(t, req.DepPath, wantPath)
	assertWorktreeListNotContains(t, req.ConsumerTop, wantPath)

	// The replace must be in the sub-module's go.mod, not the root (which has no go.mod).
	mod, err := readGoMod(req.ConsumerModDir)
	if err != nil {
		t.Fatalf("read go-pkgs/go.mod: %v", err)
	}
	if !hasReplaceForModule(mod, depModulePath, wantPath) {
		t.Fatalf("go-pkgs/go.mod missing replace %s => %s: %+v", depModulePath, wantPath, mod.Replace)
	}

	ok, err := gitignoreContainsExternal(req.ConsumerTop)
	if err != nil {
		t.Fatalf("read .gitignore: %v", err)
	}
	if !ok {
		t.Fatalf(".gitignore should contain /external")
	}
}
```