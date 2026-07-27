---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit **0** with full dashboard snapshot core (static P2 path).
- Does not hang (non-TTY must not enter Bubble Tea).

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertDashboardSnapshotCore(t, req, resp, err)
	assertAddChangesGlyph(t, resp.Stdout, true /* clean: disabled */)
	assertMergeBackDefaultSelected(t, resp.Stdout)
}
```
