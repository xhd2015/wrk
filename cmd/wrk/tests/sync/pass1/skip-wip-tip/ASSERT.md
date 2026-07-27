---
label: slow
explanation: isolated git + worktree; WIP tip skip
---

## Expected Output

```
synced: 0 into main, 0 into worktrees, 1 skipped
```

## Expected

- Exit code 0 (partial skip is success).
- Stdout is exactly the zero-action summary with skipped=1 (no detail lines).
- Stderr contains
  `warning: skip feature-login: wip commit in range (<short7> wip: half done)`
  where `<short7>` is `req.WipHashShort`.
- Main and worktree HEADs unchanged.

## Exit Code

- 0

```go
import (
	"fmt"
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

	assertHEADUnchanged(t, req.MainRepo, req.MainSHA)
	assertHEADUnchanged(t, req.WtPath, req.WtSHA)
}
```
