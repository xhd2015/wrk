## Expected

- Non-zero exit.
- Stderr includes `Error:`, branch name, both worktree paths, and refuse language
  that names commit (substring `commit` — covers `--commit` / gen-commit-msg path).
- HEAD subject equals pre-run snapshot (`req.HashToken`); no new commit.
- Both worktrees remain.

## Side Effects

- No git commit created; index may remain staged (refuse before commit).

## Errors

- Shared-branch refuse on commit path.

## Exit Code

- Non-zero

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertSharedBranchRefuseError(t, req, resp, "commit")
	assertFileExists(t, req.WtDir)
	assertFileExists(t, req.Wt2Dir)
	assertWorktreeListContains(t, req.MainRepo, req.WtDir)
	assertWorktreeListContains(t, req.MainRepo, req.Wt2Dir)

	gotSubject := strings.TrimSpace(gitOutputIsolated(t, req.RepoDir, "log", "-1", "--format=%s"))
	if gotSubject != req.HashToken {
		t.Fatalf("HEAD subject changed under refuse: got %q want %q\nstdout=%q stderr=%q",
			gotSubject, req.HashToken, resp.Stdout, resp.Stderr)
	}
}
```
