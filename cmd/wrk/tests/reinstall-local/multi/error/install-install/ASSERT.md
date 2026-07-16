## Expected

- `err` is non-nil (install×install cross-module collision).
- Error text contains `samebin` and both module identifiers (`mod-a`, `mod-b`).

## Side Effects

- None (pure discovery; may only stat/read paths).

## Exit Code

- N/A (no process).

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertMultiPlanError(t, req, resp, err)
}
```
