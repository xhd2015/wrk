## Expected

- `err` is nil.
- `ProjectPaths` contains both lib and app (order independent after normalize).
- `Modules` contains exactly:
  - lib project: `Dir="."` Path=`example.com/lib`
  - lib project: `Dir="sub"` Path=`example.com/lib/sub`
  - app project: `Dir="."` Path=`example.com/app`
- No cross or intra edges (app does not require lib in this leaf).
- `SkippedPaths` empty.

## Side Effects

- Read-only over fixtures and projects.json.

## Exit Code

- N/A (package API).

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertInventorySuccess(t, req, resp, err)
}
```
