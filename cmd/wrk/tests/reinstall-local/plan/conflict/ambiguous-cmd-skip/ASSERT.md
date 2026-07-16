## Expected

- `err` is nil.
- Items empty for bin `foo` (omitted, not Action=skip).
- One `ambiguous-cmd` warning; Paths sorted `./cmd/foo`, `./cmd/nested/foo`.

## Side Effects

- None.

## Exit Code

- N/A (no process).

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertPlanOK(t, req, resp, err)
}
```
