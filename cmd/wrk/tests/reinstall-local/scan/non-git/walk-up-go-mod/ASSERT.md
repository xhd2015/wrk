## Expected

- `err` is nil.
- `ScanRoot` is the walk-up module directory (`…/mod`), not `…/mod/nested`.
- Single module `nongit-mod` with one install item `onlybin` /
  `./cmd/onlybin`.

## Side Effects

- None (filesystem fixtures only under WorkRoot).

## Exit Code

- N/A (no process).

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertScanPlanOK(t, req, resp, err)
}
```
