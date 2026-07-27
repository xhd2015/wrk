
## Expected

- Exit code 0.
- Consumer worktree renamed to new path with new slug.
- Manual linked dep worktree at `deps/foo` exists at new path under renamed consumer wt.
- Dep main repo's worktree gitdir now points to `{newConsumerWt}/deps/foo/.git` (not stale old path).
- `git worktree list` on dep main contains new path.

```go
import (
	"path/filepath"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}

	newSlug := slugify("new slug")
	oldSlug := slugify("old slug")

	oldConsumerPath := req.WtDir
	newConsumerPath := worktreePathWithTask(req.WrkHome, "consumer", "main", wrkDate, newSlug, 0)
	newStdoutPath := strings.TrimSpace(resp.Stdout)

	assertFileNotExists(t, oldConsumerPath)
	assertFileExists(t, newConsumerPath)
	if newStdoutPath != newConsumerPath {
		t.Fatalf("stdout: expected new path %q, got %q", newConsumerPath, newStdoutPath)
	}

	oldBranch := branchNameWithTask("main", wrkDate, oldSlug, 0)
	newBranch := branchNameWithTask("main", wrkDate, newSlug, 0)
	assertBranchNotExists(t, req.MainRepo, oldBranch)
	assertBranchExists(t, req.MainRepo, newBranch)

	oldDepsPath := req.DepsLinkedWtDir
	wantDepsPath := filepath.Join(newConsumerPath, "deps", "foo")

	assertFileExists(t, wantDepsPath)
	assertGitFileIsWorktreeLink(t, wantDepsPath)

	depMain, err := readWorktreeMainRepo(wantDepsPath)
	if err != nil {
		t.Fatalf("read main repo from deps/foo .git: %v", err)
	}
	if depMain != req.DepsDepPath {
		t.Fatalf("deps/foo worktree registered under %s, want %s", depMain, req.DepsDepPath)
	}

	gotGitdir, err := readWorktreeGitdir(wantDepsPath)
	if err != nil {
		t.Fatalf("read gitdir for deps/foo worktree: %v", err)
	}
	wantGitdir := filepath.Join(wantDepsPath, ".git")
	if gotGitdir != wantGitdir {
		t.Fatalf("gitdir: expected %q, got %q", wantGitdir, gotGitdir)
	}

	oldGitdir := filepath.Join(oldDepsPath, ".git")
	if gotGitdir == oldGitdir {
		t.Fatalf("gitdir still contains stale old path %q", oldGitdir)
	}

	assertWorktreeListContains(t, depMain, wantDepsPath)
}
```