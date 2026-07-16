## Expected

- `err` is nil.
- `Releases` contains both modules:
  - `example.com/src` → Tag `v1.2.3`, Version `v1.2.3`
  - `example.com/src/sub` → Tag `sub/v0.1.0`, Version `v0.1.0`
- `Missing` is empty.

## Side Effects

- Read-only git tag listing / module scan (no new tags).

## Exit Code

- N/A (package API).

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertSourceReleasesSuccess(t, req, resp, err)
}
```
