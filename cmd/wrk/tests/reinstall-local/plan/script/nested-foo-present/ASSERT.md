## Expected

- `err` is nil.
- One item: bin `foo`, method `go-run-install`, path `./script/foo/install`,
  action `install`.

## Side Effects

- None.

## Exit Code

- N/A (no process).

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertPlanOK(t, req, resp, err)
}
```
