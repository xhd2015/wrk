## Expected

- Non-zero exit.
- Stdout empty.
- Stderr mentions mutual exclusion (and preferably `--new` and/or `--status`).
- No worktree under `{WRK_HOME}/worktrees/`.

## Errors

- `--new` cannot be combined with `--status`.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertMutexNewMode(t, resp, err, "--status")
	assertNoWorktreesCreated(t, req)
}
```
