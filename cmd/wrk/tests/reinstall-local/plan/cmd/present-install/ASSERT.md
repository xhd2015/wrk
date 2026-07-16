## Expected

- `err` is nil.
- `ModuleName` is `cmd-present`.
- Exactly one item: bin `present`, method `go-install`, path `./cmd/present`,
  action `install`.

## Side Effects

- None (pure helper; fixture files already written by Setup).

## Exit Code

- N/A (no process).

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertPlanOK(t, req, resp, err)
}
```
