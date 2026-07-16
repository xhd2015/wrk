## Expected

- `err` is nil.
- `Items` is empty — only `package main` under `cmd` is discoverable.

## Side Effects

- None.

## Exit Code

- N/A (no process).

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertPlanOK(t, req, resp, err)
}
```
