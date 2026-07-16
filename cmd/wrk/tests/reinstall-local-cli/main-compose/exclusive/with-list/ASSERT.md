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
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertMutualExclusion(t, resp)
}
```
