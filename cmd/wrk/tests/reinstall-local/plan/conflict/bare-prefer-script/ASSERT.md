## Expected

- `err` is nil.
- Exactly one item: bin `demo`, `go-run-install`, `./script/install`, install.
- No `./cmd/demo` item.
- One `prefer-script` notice with Paths `./cmd/demo`, `./script/install`.

## Side Effects

- None.

## Exit Code

- N/A (no process).

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertPlanOK(t, req, resp, err)
}
```
