## Expected

- Exit code 0.
- Scan block for main plus two appended full blocks.
- Appended order matches `worktree.ListLinked` external order (first wrk, then second).
- Each Dir follows `statusDirLine(invCwd, path)` (only formatting changes vs abs policy).
- Stderr is empty.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	assertStdoutBlocksSeparated(t, resp.Stdout, 3)

	order := externalAppendOrder(t, req.MainRepo)
	if len(order) != 2 {
		t.Fatalf("expected 2 external append entries, got %d: %v", len(order), order)
	}
	if order[0] != req.WtDir || order[1] != req.WtDir2 {
		t.Fatalf("append order: want [%q, %q], got %v", req.WtDir, req.WtDir2, order)
	}

	assertOutputExact(t, resp.Stdout, statusStdoutV2(t,
		scanStatusBlockFromCwd(t, req.RepoDir, req.MainRepo, "clean", "", true),
		appendedHealthyBlockPlain(t, req.RepoDir, req.MainRepo, req.WtDir, req.WtBranch, "clean"),
		appendedHealthyBlockPlain(t, req.RepoDir, req.MainRepo, req.WtDir2, req.WtBranch2, "clean"),
	))
}
```
