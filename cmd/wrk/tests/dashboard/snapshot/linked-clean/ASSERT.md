## Expected

- Exit **0**; full dashboard snapshot core (stages, glyphs, no create-hint, Batch would-run, no `[x]`).
- **add changes** row uses disabled glyph **`[-]`** (clean tree: no unstaged/untracked).
- **MERGE BACK** default **`[•]`**; **DONE** off when linked.
- No create under `{WRK_HOME}/worktrees/`.

## Side Effects

- None from create mode.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertDashboardSnapshotCore(t, req, resp, err)
	assertAddChangesGlyph(t, resp.Stdout, true /* clean: disabled */)
	assertMergeBackDefaultSelected(t, resp.Stdout)
}
```
