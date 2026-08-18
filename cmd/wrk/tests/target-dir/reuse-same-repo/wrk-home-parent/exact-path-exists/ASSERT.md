## Expected

- Non-zero exit.
- Stdout empty (no nested spawn path printed as success).
- Stderr indicates the **spawn/target path** is already occupied / not free
  (`exist` / `already` / `not a free` style; `wrk:` prefix).
- This is a **path collision** error, not a Policy B prompt or the legacy
  non-TTY refuse (`refusing non-interactive` must be absent).
- Existing linked worktree at `SpawnDir` remains; no nested worktree under it
  (e.g. no `{SpawnDir}/myrepo-main-{date}`).

## Errors

- Resolved intended spawn path already exists (dir/file/linked WT) — here: exact
  dump path occupied by an existing linked worktree.

## Exit Code

- non-zero

```go
import (
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero when dump spawn path already exists; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if strings.TrimSpace(resp.Stdout) != "" {
		t.Fatalf("stdout should be empty on exact-path-exists error, got %q", resp.Stdout)
	}

	// Must be path-collision style, not Policy B prompt / non-TTY refuse.
	assertNoPolicyBBanner(t, resp.Stdout+resp.Stderr)
	assertNotContains(t, resp.Stderr, "also present")
	assertNotContains(t, resp.Stderr, "refusing non-interactive")

	se := strings.ToLower(resp.Stderr)
	if !strings.Contains(se, "exist") && !strings.Contains(se, "already") && !strings.Contains(se, "not a free") {
		t.Fatalf("stderr should indicate path already exists / not free; got %q", resp.Stderr)
	}
	assertContains(t, resp.Stderr, "wrk:")
	if !strings.Contains(resp.Stderr, req.SpawnDir) && !strings.Contains(resp.Stderr, filepath.Base(req.SpawnDir)) {
		t.Fatalf("stderr should mention occupied path %q; got %q", req.SpawnDir, resp.Stderr)
	}

	assertFileExists(t, req.SpawnDir)
	assertGitFileIsWorktreeLink(t, req.SpawnDir)
	assertWorktreeListContains(t, req.TargetDir, req.SpawnDir)
	assertFileNotExists(t, filepath.Join(req.SpawnDir, "myrepo-main-"+wrkDate))
	assertFileNotExists(t, filepath.Join(req.SpawnDir, "myrepo-main-"+wrkDate+"-1"))
}
```
