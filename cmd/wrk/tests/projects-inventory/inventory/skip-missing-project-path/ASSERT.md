## Expected

- `err` is nil (missing path is soft-skip, not a hard BuildInventory error).
- `SkippedPaths` contains the non-existent registry path (normalized compare).
- `ProjectPaths` / `Modules` contain only the good project (`example.com/good`).
- No edges.

## Side Effects

- Read-only; missing path is never created.

## Exit Code

- N/A (package API).

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertInventorySuccess(t, req, resp, err)
}
```
