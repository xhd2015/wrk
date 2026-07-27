## Expected

- Exit code 0.
- Three primary blocks: main, then each ListLinked path in porcelain order
  (in-tree and out-of-tree main-owned linked).
- No `---- external ----` header (external list empty).
- Stderr is empty.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	assertStdoutBlocksSeparated(t, resp.Stdout, 3)
	assertNoExternalSectionHeader(t, resp.Stdout)

	linked := listLinkedPaths(t, req.MainRepo)
	if len(linked) != 2 {
		t.Fatalf("expected 2 ListLinked entries, got %d: %v", len(linked), linked)
	}

	primary := []string{
		scanStatusBlockFromCwd(t, req.RepoDir, req.MainRepo, "clean", "", true),
	}
	for _, p := range linked {
		branch := req.WtBranch
		statusLine := "clean"
		if resolvePath(t, p) == resolvePath(t, req.InTreeWtDir) {
			branch = req.InTreeWtBranch
		}
		primary = append(primary, primaryLinkedBlockPlain(t, req.RepoDir, req.MainRepo, p, branch, statusLine))
	}

	assertOutputExact(t, resp.Stdout, statusStdoutPrimaryExternal(t, primary, nil))
}
```
