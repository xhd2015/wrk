## Expected

- `err` is nil.
- Exactly one item for `foo`.
- Method is `go-run-install` and RelPath is `./script/foo/install` (script wins).
- No separate `./cmd/foo` item.
- One `prefer-script` notice diagnostic with sorted Paths
  `./cmd/foo`, `./script/foo/install`.

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
