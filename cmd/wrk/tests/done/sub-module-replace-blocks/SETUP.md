# Scenario

**Feature**: wrk --done errors when a sub-module (not the main go.mod) has a filesystem replace

```
# main go.mod clean; sub/go.mod has replace => ./local-foo
# scan must find the sub-module and block --done, not just check the main go.mod
consumer wt (main clean + sub has replace) -> wrk --done -> guard error before merge-back
```

## Steps

1. Create main repo on `main` with a clean top-level `go.mod` (no replace).
2. Add a sub-module at `sub/go.mod` carrying `replace example.com/foo => ./local-foo`.
3. Commit both, then create a linked worktree via `wrk`.
4. Run `wrk --done` from the linked worktree root.

## Expected (correct) behavior

The local-replace guard scans **every** Go module under the checkout
(`gotool/mod/scan.Scan`), not just the nearest top-level `go.mod`. The
sub-module's `replace => ./local-foo` is a local filesystem replace, so `--done`
must block before merge-back, leaving the worktree in place.

## Bug

Current behavior: `runDone` calls `findGoModDir` which resolves the clean
top-level `go.mod`, then `HasLocalFilesystemReplace` checks only that one
module. The sub-module replace is never inspected, so `--done` proceeds,
merge-back removes the worktree, and the sub-module's local replace is merged
into main unchecked.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)

	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)

	// Clean top-level go.mod — no replace here. The guard must look deeper.
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/myrepo\n\ngo 1.22\n")

	// Sub-module with a local filesystem replace. Its target need not exist;
	// the guard classifies by path prefix, and scan parses go.mod directly.
	mkdirAll(t, filepath.Join(mainRepo, "sub"))
	writeFile(t, filepath.Join(mainRepo, "sub", "go.mod"),
		"module example.com/myrepo/sub\n\ngo 1.22\n\nreplace example.com/foo => ./local-foo\n")
	runGitIsolated(t, mainRepo, "add", "go.mod", "sub/go.mod")
	// --no-verify bypasses the project's go-no-local-replace pre-commit hook,
	// which would otherwise reject committing sub/go.mod's local replace. This
	// test exercises wrk --done's guard, not the hook.
	runGitIsolated(t, mainRepo, "commit", "--no-verify", "-m", "add main + sub module")

	wtDir := runWrkFrom(t, req, mainRepo)
	req.WtDir = wtDir
	req.WtBranch = branchName("main", wrkDate, 0)

	req.RepoDir = wtDir
	req.Args = []string{"--done"}
	return nil
}
```
