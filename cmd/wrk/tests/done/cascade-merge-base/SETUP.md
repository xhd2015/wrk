# Scenario

**Feature**: wrk --done cascade removes an external dep worktree without crashing on merge-base

```
# wrk --dep spawns external dep wt as a worktree of the CONSUMER repo, its branch
# a fetched dep commit. --done cascades MergeBack over external/* — that must
# remove the dep wt, not crash comparing unrelated histories.
consumer wt + external/dep wt -> wrk --done -> cascade removes dep wt (no merge-base error)
```

## Steps

1. Create consumer main repo on `main` with `go.mod` requiring `example.com/dep`.
2. `wrk` creates the consumer linked worktree.
3. `wrk --dep <depRepo>` spawns `external/mydep-main-{date}` (a worktree of the
   consumer repo whose branch is a fetched dep commit).
4. Drop the consumer's `replace => ./external/...` so the local-replace guard is
   NOT the thing under test — the cascade is the only moving part.
5. Run `wrk --done` from the consumer worktree.

## Expected (correct) behavior

The cascade merges-back / removes each `external/*` linked worktree. The dep
worktree's branch is a fetched dep commit living in the consumer repo and shares
no ancestry with consumer `main`; there is nothing to merge back, so the cascade
must simply remove it. `--done` must not crash with `failed to find merge base`.

## Bug

`createExternalWorktree` (`go-pkgs/wrkcli/run.go`) builds the dep worktree via
`git -C consumerMain worktree add -b <branch> <ext> wrk-dep-<dep>/<depBranch>` —
so the worktree's common git dir is the **consumer** repo and its branch points
at a **fetched dep commit**. When `runDone` cascades `MergeBack` over it,
`ReadMainRepo(ext)` returns the consumer main repo and `CompareBranches` runs
`git -C consumerMain merge-base <dep-branch> HEAD`. The dep-branch commit and
consumer `HEAD` come from different repositories with no shared history, so
`git merge-base` exits 1 → `failed to find merge base: exit status 1`. The error
propagates and aborts `--done` before the dep worktree is removed.

```go
import (
	"os/exec"
	"path/filepath"
)

const cascadeMbDepModule = "example.com/dep"

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	if _, err := exec.LookPath("go"); err != nil {
		t.Fatalf("go not found: %v", err)
	}

	mainRepo := filepath.Join(req.WorkRoot, "consumer")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/consumer\n\ngo 1.22\n")
	runGoMod(t, mainRepo, "edit", "-require="+cascadeMbDepModule+"@v0.0.0")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add consumer go.mod")

	wtDir := runWrkFrom(t, req, mainRepo)
	req.WtDir = wtDir
	req.WtBranch = branchName("main", wrkDate, 0)

	depRepo := filepath.Join(req.WorkRoot, "mydep")
	initGitRepoOnMain(t, depRepo)
	writeFile(t, filepath.Join(depRepo, "go.mod"), "module "+cascadeMbDepModule+"\n\ngo 1.22\n")
	writeFile(t, filepath.Join(depRepo, "dep.go"), "package dep\n")
	runGitIsolated(t, depRepo, "add", "go.mod", "dep.go")
	runGitIsolated(t, depRepo, "commit", "-m", "add module")

	externalPath := runWrkWithArgs(t, req, wtDir, "--dep", depRepo)
	req.ExternalWtDir = externalPath

	// Drop the consumer's local replace so the local-replace guard is not what
	// fails — the cascade merge-base crash is the sole behavior under test.
	runGoMod(t, wtDir, "edit", "-dropreplace="+cascadeMbDepModule)

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
