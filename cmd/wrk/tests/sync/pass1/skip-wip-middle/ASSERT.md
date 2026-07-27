---
label: slow
explanation: isolated git + worktree; WIP middle-of-range skip
---

## Expected Output

```
synced: 0 into main, 0 into worktrees, 1 skipped
```

## Expected

- Exit code 0.
- Stdout is exactly zero-action summary with skipped=1 (no detail lines).
- Stderr contains warning naming the **first** WIP in range:
  `warning: skip feature-login: wip commit in range (<short7 of middle wip> wip: half done)`
  — not the clean tip subject.
- Main and worktree HEADs unchanged.

## Exit Code

- 0

```go
import (
	"fmt"
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantOut := buildSyncStdout(nil, 0, 0, 1, false)
	assertOutputExact(t, resp.Stdout, syncStdoutV2(wantOut))

	wantWarn := fmt.Sprintf(
		"warning: skip feature-login: wip commit in range (%s %s)",
		req.WipHashShort, req.WipSubject,
	)
	assertContains(t, resp.Stderr, wantWarn)
	// Named wip subject must be the middle commit, not the clean tip.
	if strings.Contains(resp.Stderr, "feat: clean tip") {
		t.Fatalf("warning should name first WIP subject, not clean tip; stderr=%q", resp.Stderr)
	}

	assertHEADUnchanged(t, req.MainRepo, req.MainSHA)
	assertHEADUnchanged(t, req.WtPath, req.WtSHA)
}
```
