# Scenario

**Feature**: wrk --done cascades merge-back to both external/* and non-external linked worktrees

```
# consumer wt + external/dep wt (from --dep) + manual deps/foo linked wt -> wrk --done
# scan_repo.Scan discovers both nested linked wts; cascade removes each before consumer merge-back
consumer wt + external wt + deps/foo wt -> wrk --done -> both removed, consumer exit 0
```

## Steps

1. Create consumer main repo with `go.mod` requiring `example.com/dep`.
2. `wrk` creates the consumer linked worktree.
3. `wrk --dep <depRepo>` spawns `external/mydep-main-{date}`.
4. Create a second dep main repo and run `git worktree add` into `{consumerWt}/deps/foo`.
5. Drop the consumer's `replace => ./external/...` so the local-replace guard does not block.
6. Run `wrk --done` from the consumer worktree.

```go
import (
	"os/exec"
	"path/filepath"
)

const cascadeBothDepModule = "example.com/dep"
const cascadeBothFooModule = "example.com/foodep"

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	if _, err := exec.LookPath("go"); err != nil {
		t.Fatalf("go not found: %v", err)
	}

	mainRepo := filepath.Join(req.WorkRoot, "consumer")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/consumer\n\ngo 1.22\n")
	runGoMod(t, mainRepo, "edit", "-require="+cascadeBothDepModule+"@v0.0.0")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add consumer go.mod")

	wtDir := runWrkFrom(t, req, mainRepo)
	req.WtDir = wtDir
	req.WtBranch = branchName("main", wrkDate, 0)

	depRepo := filepath.Join(req.WorkRoot, "mydep")
	req.DepPath = depRepo
	initGitRepoOnMain(t, depRepo)
	writeFile(t, filepath.Join(depRepo, "go.mod"), "module "+cascadeBothDepModule+"\n\ngo 1.22\n")
	writeFile(t, filepath.Join(depRepo, "dep.go"), "package dep\n")
	runGitIsolated(t, depRepo, "add", "go.mod", "dep.go")
	runGitIsolated(t, depRepo, "commit", "-m", "add module")

	externalPath := runWrkWithArgs(t, req, wtDir, "--dep", depRepo)
	req.ExternalWtDir = externalPath

	fooDepRepo := filepath.Join(req.WorkRoot, "foodep")
	req.DepsDepPath = fooDepRepo
	initGitRepoOnMain(t, fooDepRepo)
	writeFile(t, filepath.Join(fooDepRepo, "go.mod"), "module "+cascadeBothFooModule+"\n\ngo 1.22\n")
	writeFile(t, filepath.Join(fooDepRepo, "foo.go"), "package foo\n")
	runGitIsolated(t, fooDepRepo, "add", "go.mod", "foo.go")
	runGitIsolated(t, fooDepRepo, "commit", "-m", "add module")

	depsWtDir := filepath.Join(wtDir, "deps", "foo")
	req.DepsLinkedWtDir = depsWtDir
	fooBranch := branchName("main", wrkDate, 0)
	runGitIsolated(t, fooDepRepo, "worktree", "add", "-b", fooBranch, depsWtDir)

	runGoMod(t, wtDir, "edit", "-dropreplace="+cascadeBothDepModule)

	// Keep consumer wt clean after external/ + deps/ worktrees exist.
	// Stage all --dep dirt (go.mod, .gitignore, go.sum if present) so --done can remove.
	writeFile(t, filepath.Join(wtDir, ".gitignore"), "/external\n/deps\n")
	runGitIsolated(t, wtDir, "add", "-A")
	runGitIsolated(t, wtDir, "commit", "-m", "drop replace and ignore nested worktrees")

	req.RepoDir = wtDir
	// Default auto-yes: no --confirm / --confirm-from-stdin required for own-wt ahead.
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