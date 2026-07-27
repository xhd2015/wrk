## Expected

- Exit code 0.
- Stdout (trimmed) equals `{consumerTop}/external/mydep-main-{WRK_DATE}`.
- External path exists as a linked git worktree.
- **The external worktree is a worktree of the DEP repo**: its `.git` gitdir
  points into `<dep-main>/.git/worktrees/`, not the consumer's.
- `git worktree list` in the dep main repo contains the external path.
- `git worktree list` in the consumer repo does NOT contain the external path
  (the consumer repo merely hosts the working tree on disk; it does not own it).

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantPath := externalWorktreePath(req.ConsumerTop, "mydep", "main", 0)
	req.ExternalWtDir = wantPath
	assertStdoutExactPath(t, resp.Stdout, wantPath)

	assertFileExists(t, wantPath)
	assertGitFileIsWorktreeLink(t, wantPath)

	// The external worktree is a worktree of the DEP repo, so its .git gitdir
	// must resolve to the dep's main repo, not the consumer's. In the test
	// layout the dep is a main checkout, so dep main == req.DepPath.
	depMain := req.DepPath
	gotMain, err := readWorktreeMainRepo(wantPath)
	if err != nil {
		t.Fatalf("read external worktree .git: %v", err)
	}
	if gotMain != depMain {
		t.Fatalf("external worktree .git gitdir resolves to main repo %s, want dep main %s\n"+
			"(the dep worktree is wrongly registered under the consumer repo)", gotMain, depMain)
	}

	// The dep repo owns the worktree...
	assertWorktreeListContains(t, depMain, wantPath)
	// ...and the consumer repo does NOT.
	assertWorktreeListNotContains(t, req.ConsumerTop, wantPath)
}
```
