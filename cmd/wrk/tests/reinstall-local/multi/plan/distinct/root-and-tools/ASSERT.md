## Expected

- `err` is nil.
- Two modules in lex ModuleRoot order: `.../root` then `.../root/tools`.
- Root module installs `rootbin` via `./cmd/rootbin`; tools module installs
  `toolbin` via `./cmd/toolbin`.

## Side Effects

- None (pure discovery).

## Exit Code

- N/A (no process).

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertMultiPlanOK(t, req, resp, err)
}
```
