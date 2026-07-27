## Expected

- `err` is nil.
- `ModuleName` is `demo`.
- One item: bin `demo` (module basename, not `install`/`script`), method
  `go-run-install`, path `./script/install`, action `install`.

## Side Effects

- None.

## Exit Code

- N/A (no process).

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertPlanOK(t, req, resp, err)
}
```
