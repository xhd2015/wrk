## Expected

- Non-zero exit.
- Stderr includes `Error:`, branch name, both worktree paths, refuse language naming commit.
- HEAD subject equals pre-run snapshot; no new commit.

## Side Effects

- No git commit created; index may remain staged (refuse before commit).

## Errors

- Shared-branch refuse on manual commit path.

## Exit Code

- Non-zero

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertSharedBranchRefuseError(t, req, resp)

	gotSubject := gitHEADSubject(t, req.RepoDir)
	if gotSubject != req.HEADSubject {
		t.Fatalf("HEAD subject changed under refuse: got %q want %q\nstdout=%q stderr=%q",
			gotSubject, req.HEADSubject, resp.Stdout, resp.Stderr)
	}
}
```
