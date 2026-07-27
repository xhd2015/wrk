## Expected

- `err` is nil.
- Items empty for bin `foo` (omitted, not Action=skip).
- One `ambiguous-script` warning; Paths sorted
  `./script/foo/install`, `./script/x/foo/install`.

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
