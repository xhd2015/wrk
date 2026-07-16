## Expected

- `err` is nil.
- `ScanRoot` is the git toplevel (`…/repo`), not `pkg/sub` and not a walk-up
  to a single nested module only.
- Two modules in lex ModuleRoot order: `…/repo` then `…/repo/tools`.
- Root installs `rootbin` via `./cmd/rootbin`; tools installs `toolbin` via
  `./cmd/toolbin`.

## Side Effects

- None beyond fixture git writes in WorkRoot (no production writes).

## Exit Code

- N/A (no process).

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertScanPlanOK(t, req, resp, err)
}
```
