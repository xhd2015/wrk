## Expected

- Exit code 0.
- Consumer worktree renamed to new path with new slug.
- External dep worktree's gitdir in dep main repo now points to new path (not old path).
- External dep worktree's `.git` file still points to correct dep main repo.

```go
import (
	"path/filepath"
	"strings"
	"testing"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	if err != nil {
		t.Fatal(err)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}

	newSlug := slugify("new slug")
	oldSlug := slugify("old slug")

	// Consumer worktree renamed
	oldConsumerPath := req.WtDir
	newConsumerPath := worktreePathWithTask(req.WrkHome, "consumer", "main", wrkDate, newSlug, 0)
	newStdoutPath := strings.TrimSpace(resp.Stdout)

	assertFileNotExists(t, oldConsumerPath)
	assertFileExists(t, newConsumerPath)
	if newStdoutPath != newConsumerPath {
		t.Fatalf("stdout: expected new path %q, got %q", newConsumerPath, newStdoutPath)
	}

	// Branch renamed in consumer main
	oldBranch := branchNameWithTask("main", wrkDate, oldSlug, 0)
	newBranch := branchNameWithTask("main", wrkDate, newSlug, 0)
	assertBranchNotExists(t, req.MainRepo, oldBranch)
	assertBranchExists(t, req.MainRepo, newBranch)

	// External dep worktree: recompute expected new path
	oldExtPath := req.ExternalWtDir
	wantExtPath := filepath.Join(newConsumerPath, "external", filepath.Base(oldExtPath))

	// External dep worktree directory moved with consumer
	assertFileExists(t, wantExtPath)
	assertGitFileIsWorktreeLink(t, wantExtPath)

	// Dep main repo still owns the worktree
	depMain, err := readWorktreeMainRepo(wantExtPath)
	if err != nil {
		t.Fatalf("read main repo from ext .git: %v", err)
	}
	if depMain != req.DepPath {
		t.Fatalf("external worktree registered under %s, want %s", depMain, req.DepPath)
	}

	// gitdir in dep's main repo should now point to new external path + "/.git"
	gotGitdir, err := readWorktreeGitdir(wantExtPath)
	if err != nil {
		t.Fatalf("read gitdir for ext worktree: %v", err)
	}
	wantGitdir := filepath.Join(wantExtPath, ".git")
	if gotGitdir != wantGitdir {
		t.Fatalf("gitdir: expected %q, got %q", wantGitdir, gotGitdir)
	}

	// gitdir should NOT be the old path
	oldGitdir := filepath.Join(oldExtPath, ".git")
	if gotGitdir == oldGitdir {
		t.Fatalf("gitdir still contains stale old path %q", oldGitdir)
	}

	// worktree listed in dep main repo at new path
	assertWorktreeListContains(t, depMain, wantExtPath)
}
```
