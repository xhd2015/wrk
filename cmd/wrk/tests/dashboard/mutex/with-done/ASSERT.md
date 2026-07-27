## Expected

- Non-zero exit.
- Stdout empty.
- Stderr mentions mutual exclusion (and preferably `--new` and/or `--done`).
- No worktree under `{WRK_HOME}/worktrees/`.

## Errors

- `--new` cannot be combined with `--done`.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertMutexNewMode(t, resp, err, "--done")
	assertNoWorktreesCreated(t, req)
}
```
