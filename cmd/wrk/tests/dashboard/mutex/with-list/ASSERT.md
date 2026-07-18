## Expected

- Non-zero exit.
- Stdout empty.
- Stderr mentions mutual exclusion (and preferably `--new` and/or `--list`).
- No worktree under `{WRK_HOME}/worktrees/`.

## Errors

- `--new` cannot be combined with `--list`.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertMutexNewMode(t, resp, err, "--list")
	assertNoWorktreesCreated(t, req)
}
```
