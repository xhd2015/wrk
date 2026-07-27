## Expected

- `err` is nil.
- `ProjectPaths` is empty.
- `Modules` is empty.
- `CrossEdges` and `IntraEdges` are empty.
- `SkippedPaths` is empty.

## Side Effects

- None beyond reading absent/empty registry (no writes).

## Exit Code

- N/A (package API).

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertInventorySuccess(t, req, resp, err)
}
```
