---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit **0**.
- Full P2 snapshot core (dashboard identity, Pre/Main/After/Batch, stages, glyphs; **no** create-hint).
- add changes disabled `[-]` on clean tree.
- MERGE BACK default when linked.
- No create under `WRK_HOME/worktrees`.

## Side Effects

- None.

## Exit Code

- 0

```go
func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertDashboardSnapshotCore(t, req, resp, err)
	assertAddChangesGlyph(t, resp.Stdout, true /* clean */)
	assertMergeBackDefaultSelected(t, resp.Stdout)
}
```
