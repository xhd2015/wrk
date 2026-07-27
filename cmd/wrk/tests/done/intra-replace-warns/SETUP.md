# Scenario

**Feature**: wrk --done warns (and proceeds) for an intra-repo filesystem replace

```
# consumer go.mod has replace example.com/foo => ./submod; ./submod is a real
# nested module sharing the consumer's git toplevel -> intra-repo -> WARN + proceed
consumer wt (replace => ./submod, same repo) -> wrk --done -> warn on stderr, merge-back runs
```

## Steps

1. Create main repo on `main` with a consumer `go.mod` whose
   `replace example.com/foo => ./submod` points at a real nested module
   (`submod/go.mod`, `module example.com/foo`) committed in the same repo. This
   stands in for the user's `replace github.com/xhd2015/ai-critic => ../../`:
   both resolve to an existing directory sharing the consumer's
   `git rev-parse --show-toplevel`, i.e. intra-repo.
2. Create a linked worktree via `wrk`.
3. Run `wrk --done` from the linked worktree (default, no flag).

## Expected (correct) behavior

The guard classifies the `./submod` replace as **intra-repo** (target exists and
shares the consumer's toplevel). Under the default lenient guard it **warns to
stderr and proceeds** (exit 0), so merge-back runs and removes the worktree.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	skipIfNoGit(t)

	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)

	// Real nested module that the replace points at — exists + same toplevel.
	mkdirAll(t, filepath.Join(mainRepo, "submod"))
	writeFile(t, filepath.Join(mainRepo, "submod", "go.mod"), "module example.com/foo\n\ngo 1.21\n")

	// Consumer module with an intra-repo filesystem replace.
	writeFile(t, filepath.Join(mainRepo, "go.mod"),
		"module example.com/consumer\n\ngo 1.22\n\nreplace example.com/foo => ./submod\n")
	runGitIsolated(t, mainRepo, "add", "go.mod", "submod")
	// --no-verify bypasses the project's go-no-local-replace pre-commit hook,
	// which would otherwise reject committing an intra-repo replace. This test
	// exercises wrk --done's guard, not the hook; the worktree must be clean so
	// merge-back can proceed after the warn.
	runGitIsolated(t, mainRepo, "commit", "--no-verify", "-m", "add consumer with intra-repo replace")

	wtDir := runWrkFrom(t, req, mainRepo)
	req.WtDir = wtDir
	req.WtBranch = branchName("main", wrkDate, 0)

	req.RepoDir = wtDir
	req.Args = []string{"--done"}
	return nil
}
```
