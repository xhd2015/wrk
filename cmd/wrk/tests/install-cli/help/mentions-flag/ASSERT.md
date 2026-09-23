
## Expected

- Exit code 0.
- Help stdout contains `--install` and the required-arg/`--main` notes.
- No event is appended for help (not asserted here; covered by existing trees).

## Side Effects

- None (help only).

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	assertContains(t, resp.Stdout, "--install")
	assertContains(t, resp.Stdout, "at least one name")
	assertContains(t, resp.Stdout, "--reinstall-local")
}
```
