## Expected

- `err` is nil (skip-only duplicate is not a hard collision).
- Two modules in lex ModuleRoot order (`mod-a` then `mod-b`).
- Each lists one item: BinName `sharedbin`, Method `go-install`, Action `skip`.

## Side Effects

- None (pure discovery).

## Exit Code

- N/A (no process).

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertMultiPlanOK(t, req, resp, err)
}
```
