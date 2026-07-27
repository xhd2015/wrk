## Expected

- Exit code 0.
- Primary: main plus two out-of-tree full blocks in `worktree.ListLinked` porcelain order.
- No `---- external ----` header (both linked are main-owned primary).
- Each Dir follows `statusDirLine(invCwd, path)`.
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

	order := listLinkedPaths(t, req.MainRepo)
	if len(order) != 2 {
		t.Fatalf("expected 2 ListLinked entries, got %d: %v", len(order), order)
	}
	if resolvePath(t, order[0]) != resolvePath(t, req.WtDir) ||
		resolvePath(t, order[1]) != resolvePath(t, req.WtDir2) {
		t.Fatalf("ListLinked order: want [%q, %q], got %v", req.WtDir, req.WtDir2, order)
	}

	assertOutputExact(t, resp.Stdout, statusStdoutPrimaryExternal(t,
		[]string{
			scanStatusBlockFromCwd(t, req.RepoDir, req.MainRepo, "clean", "", true),
			appendedHealthyBlockPlain(t, req.RepoDir, req.MainRepo, req.WtDir, req.WtBranch, "clean"),
			appendedHealthyBlockPlain(t, req.RepoDir, req.MainRepo, req.WtDir2, req.WtBranch2, "clean"),
		},
		nil,
	))
}
```
