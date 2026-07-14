# Scenario

**Feature**: wrk done blocks absolute cross-worktree filesystem replace

```
# linked wt go.mod has absolute replace to main-checkout submod (sibling worktree path)
# different ShowToplevel -> extra-repo -> block (same as stale external absolute paths)
WorkRoot -> wrk done wtDir -> guard blocks, worktree remains
```

## Steps

1. Create main repo with nested submod module and consumer go.mod using
   replace example.com/foo => ./submod (committed on main).
2. Create a linked worktree via wrk.
3. Rewrite the worktree replace to an absolute path pointing at mainRepo/submod.
4. Commit on the worktree branch so the tree is clean.
5. Run wrk with done flag and wtDir from WorkRoot (shell cwd outside the worktree).

## Expected (correct) behavior

The absolute path targets another checkout (main repo path) with a different
ShowToplevel than the linked worktree, so the guard classifies it as extra-repo
and blocks done. The linked worktree is not removed.

```go
import (
	"os/exec"
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
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

	submodAbs := filepath.Join(mainRepo, "submod")
	runGoMod(t, wtDir, "edit", "-replace=example.com/foo="+submodAbs)
	runGitIsolated(t, wtDir, "add", "go.mod")
	runGitIsolated(t, wtDir, "commit", "--no-verify", "-m", "absolute cross-worktree replace")

	req.RepoDir = req.WorkRoot
	req.Args = []string{"--done", wtDir}
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