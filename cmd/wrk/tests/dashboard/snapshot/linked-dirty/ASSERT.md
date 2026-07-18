## Expected

- Exit **0**.
- Stdout is a **dashboard** snapshot (not a create path).
- Glyphs use **`[•]`** / **`[ ]`** / **`[-]`**; **no** `[x]` / `[X]`.
- Section labels **Pre**, **Main**, **After** visible; Batch **would run** present.
- Stages present: **add changes**, **gen-commit-msg**, nested **agent-runner** (default **commandcode**), **commit**, **MERGE BACK** then **DONE**, **sync**, **tag-next**, **push**, **reinstall-local**.
- **add changes** enabled for dirty tree: glyph **`[•]`** or **`[ ]`** (not only `[-]`).
- **MERGE BACK** default-selected **`[•]`**; **DONE** not selected.
- **No** create-hint (`create a worktree` / `hint: create…` / `wrk --new` tip).
- No worktree created under `{WRK_HOME}/worktrees/` by this bare invocation.

## Side Effects

- None from create mode.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertDashboardSnapshotCore(t, req, resp, err)
	assertAddChangesGlyph(t, resp.Stdout, false /* dirty: not disabled-only */)
	assertMergeBackDefaultSelected(t, resp.Stdout)
}
```
