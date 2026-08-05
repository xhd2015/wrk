## Expected Output

One absolute path for the linked worktree (cwd), trailing newline.

## Expected

- Exit code 0.
- Stdout is the linked worktree absolute path plus trailing `\n`.
- Stderr is empty.
- Product must consult cwd main when origin matches even if not yet in projects.json
  (auto-record during the same invocation is allowed and does not weaken the contract).

## Side Effects

- Read-only location lookup for paths; projects.json may gain an auto-record entry.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertStdoutExactPath(t, resp.Stdout, resolvePath(t, req.WtDir))
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
}
```
