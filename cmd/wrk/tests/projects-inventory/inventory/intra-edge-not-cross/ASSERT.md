## Expected

- `err` is nil.
- Modules include root and nested sub under the same monorepo project.
- `IntraEdges` contains exactly the root→sub require
  (`example.com/mono` → `example.com/mono/sub`, owner = monorepo path).
- `CrossEdges` is empty (same-project requires are not cross).

## Side Effects

- Read-only.

## Exit Code

- N/A (package API).

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertInventorySuccess(t, req, resp, err)
}
```
