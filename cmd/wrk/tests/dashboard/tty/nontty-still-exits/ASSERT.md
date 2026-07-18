## Expected

- Exit **0** with full dashboard snapshot core (static P2 path).
- Does not hang (non-TTY must not enter Bubble Tea).

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertDashboardSnapshotCore(t, req, resp, err)
	assertAddChangesGlyph(t, resp.Stdout, true /* clean: disabled */)
	assertMergeBackDefaultSelected(t, resp.Stdout)
}
```
