## Expected

- `err` is nil.
- Inventory includes app and lib projects/modules.
- `CrossEdges` has exactly one edge:
  - consumer project = app path
  - consumer module = `example.com/app`
  - dep path = `example.com/lib`
  - dep version = `v1.0.0`
  - owner project = lib path
- `IntraEdges` is empty (require is not same-project).

## Side Effects

- Read-only.

## Exit Code

- N/A (package API).

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertInventorySuccess(t, req, resp, err)
}
```
