---
label: e2e
explanation: product binary CLI integration (process boundary)
---

## Expected

- Exit **0**.
- Full dashboard snapshot core (identity, Pre/Main/After/Batch, stages, glyphs; **no** create-hint).
- **No** create under `{WRK_HOME}/worktrees/`.
- Soft: **DONE** / **MERGE BACK** may appear as disabled `[-]` when cwd is main (rows must still appear).

## Side Effects

- None from create mode.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/session"
	"strings"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertDashboardSnapshotCore(t, req, resp, err)
	// Soft preference: main may disable DONE / MERGE BACK.
	out := resp.Stdout
	doneLn := lineContaining(out, "DONE")
	mergeLn := lineContaining(out, "MERGE BACK")
	if mergeLn == "" {
		for _, line := range strings.Split(out, "\n") {
			if strings.Contains(strings.ToLower(line), "merge back") ||
				strings.Contains(strings.ToLower(line), "merge-back") {
				mergeLn = line
				break
			}
		}
	}
	// Soft: if both rows exist and neither is disabled, still pass
	// (product may keep them selectable). Core already requires the labels.
	_ = doneLn
	_ = mergeLn
}
```
