---
label: e2e, tty
explanation: requires `script` fake TTY so Policy B would prompt today; dump parent must skip it
---

## Expected

- Exit code 0.
- Stdout is the **new** exact spawn path (`req.SpawnDir`) plus trailing `\n` — not the dump sibling.
  (`script(1)` may prefix control characters; when stdout is otherwise exactly the path,
  lock it with `assertStdoutExactPath`.)
- New path exists as a live linked worktree of source and is listed on `myrepo`.
- Dump sibling remains on disk and remains listed.
- Combined stdout+stderr does **not** contain Policy B tokens: `would reuse`,
  `skip creating`, `also present`, `already has a linked worktree`,
  `refusing non-interactive`.
- Path occupancy is locked; branch name is not (preferred `main-{date}` is taken,
  so create may suffix the branch as today).

## Side Effects

- New linked worktree at the missing dump name.
- Existing dump sibling unchanged.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantNew := req.SpawnDir
	got := strings.TrimSpace(resp.Stdout)
	if got != wantNew && !strings.Contains(resp.Stdout, wantNew) {
		t.Fatalf("stdout should be/include new dump path %q; stdout=%q stderr=%q", wantNew, resp.Stdout, resp.Stderr)
	}
	if got == wantNew {
		assertStdoutExactPath(t, resp.Stdout, wantNew)
	}
	if got == req.WtDir {
		t.Fatalf("stdout must not be dump sibling %q", req.WtDir)
	}

	assertFileExists(t, wantNew)
	assertGitFileIsWorktreeLink(t, wantNew)
	assertWorktreeListContains(t, req.TargetDir, wantNew)

	assertFileExists(t, req.WtDir)
	assertGitFileIsWorktreeLink(t, req.WtDir)
	assertWorktreeListContains(t, req.TargetDir, req.WtDir)

	combined := resp.Stdout + resp.Stderr
	assertNoPolicyBBanner(t, combined)
	assertNotContains(t, combined, "also present")
}
```
