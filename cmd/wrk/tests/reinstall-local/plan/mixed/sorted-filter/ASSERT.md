## Expected

- `err` is nil.
- Exactly three items in lexicographic BinName order: `alpha`, `mid`, `zed`.
- `alpha` and `zed` are `install`; `mid` is `skip`.
- Methods/paths: cmd for alpha/zed, script for mid.

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
