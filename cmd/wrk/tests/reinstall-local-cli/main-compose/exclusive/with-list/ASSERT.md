## Expected

- Non-zero exit code.
- Stderr mentions mutual exclusion (or unexpected arguments).
- Stdout is empty.
- Compose of `--main` + `--reinstall-local` does **not** allow stacking `--list`.

## Errors

- `--main --reinstall-local` cannot be combined with `--list`.

## Exit Code

- Non-zero

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertMutualExclusion(t, resp)
}
```
