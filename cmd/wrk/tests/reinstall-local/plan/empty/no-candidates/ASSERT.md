## Expected

- `err` is nil.
- `ModuleName` is `empty-mod`.
- `Items` is empty (zero install actions; plan ok).

## Side Effects

- None.

## Exit Code

- N/A (no process).

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertPlanOK(t, req, resp, err)
}
```
