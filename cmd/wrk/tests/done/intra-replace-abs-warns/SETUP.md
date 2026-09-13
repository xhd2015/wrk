# Scenario

**Feature**: wrk --done warns (and proceeds) for an absolute intra-repo filesystem replace

```
# linked wt go.mod has absolute replace to the same worktree's submod
# same ShowToplevel -> absolute intra-repo -> WARN + proceed
consumer wt (replace => /abs/.../submod, same repo) -> wrk --done -> warn, merge-back runs
```

## Steps

1. Create main repo with nested submod module and consumer go.mod using
   relative `replace example.com/foo => ./submod` (committed on main).
2. Create a linked worktree via `wrk`.
3. Rewrite the worktree replace to an absolute path pointing at
   `wtDir/submod` (same checkout toplevel as the consumer).
4. Commit on the worktree branch so the tree is clean.
5. Run `wrk --done` from the linked worktree (default, no flag).

## Expected (correct) behavior

The absolute path targets a directory inside the same worktree toplevel, so the
guard classifies it as **intra-repo**. Because the literal `NewPath` is absolute,
the default lenient guard **warns to stderr and proceeds** (exit 0); merge-back
runs and removes the worktree.

```go
import (
	"os/exec"
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	skipIfNoGit(t)
	if _, err := exec.LookPath("go"); err != nil {
		t.Fatalf("go not found: %v", err)
	}

	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)

	mkdirAll(t, filepath.Join(mainRepo, "submod"))
	writeFile(t, filepath.Join(mainRepo, "submod", "go.mod"), "module example.com/foo\n\ngo 1.21\n")
	writeFile(t, filepath.Join(mainRepo, "go.mod"),
		"module example.com/consumer\n\ngo 1.22\n\nreplace example.com/foo => ./submod\n")
	runGitIsolated(t, mainRepo, "add", "go.mod", "submod")
	runGitIsolated(t, mainRepo, "commit", "--no-verify", "-m", "add consumer with intra-repo replace")

	wtDir := runWrkFrom(t, req, mainRepo)
	req.WtDir = wtDir
	req.WtBranch = branchName("main", wrkDate, 0)

	// Absolute path into the same linked worktree (same ShowToplevel).
	submodAbs := filepath.Join(wtDir, "submod")
	runGoMod(t, wtDir, "edit", "-replace=example.com/foo="+submodAbs)
	runGitIsolated(t, wtDir, "add", "go.mod")
	runGitIsolated(t, wtDir, "commit", "--no-verify", "-m", "absolute same-toplevel replace")

	req.RepoDir = wtDir
	req.Args = []string{"--done"}
	return nil
}

func runGoMod(t *testing.T, dir string, args ...string) {
	t.Helper()
	cmd := exec.Command("go", append([]string{"mod"}, args...)...)
	cmd.Dir = dir
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("go mod %v: %v\n%s", args, err, out)
	}
}
```
