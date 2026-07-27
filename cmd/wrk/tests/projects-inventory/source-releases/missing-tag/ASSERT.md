## Expected

- `err` is nil (partial missing tags do not fail the whole call).
- `Releases` contains only root: ModulePath `example.com/src`, Tag/Version `v1.0.0`.
- `Missing` contains `example.com/src/sub` (no numeric tag for nested module).
- Nested module must **not** appear in Releases.

## Side Effects

- Read-only.

## Exit Code

- N/A (package API).

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertSourceReleasesSuccess(t, req, resp, err)
}
```
