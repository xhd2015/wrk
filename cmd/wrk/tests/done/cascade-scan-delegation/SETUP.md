# Scenario

**Feature**: `wrk --done` cascade discovery delegates to `scan_repo.Scan` with
`ListWorktrees: true` and base-path filter; cascade targets include linked paths
from top-level scan rows **and** inner `Repo.Worktrees`

```
# cascade discovery (product contract after P2)
consumerTop = ShowToplevel(cwd)
  -> scan_repo.Scan(Roots: [consumerTop], ListWorktrees: true)
  -> collect cascade targets from:
       - top-level RepoTypeWorktree rows under consumer (e.g. external/*, deps/*)
       - inner repo.Worktrees on main rows (linked, under consumer, not self)
  -> skip consumer checkoutRoot; nested RepoTypeMain under consumer → hard Error: (D1)
  -> dirty preflight only on cascade targets + own (foreign outside base must not appear)

# foreign isolation (P1 filter + P2 wiring regression)
WorkRoot/other/external/agent-pro  # dirty main OUTSIDE consumerTop
consumer linked wt clean, zero nested under consumer
  -> wrk --done | --done --dry-run
  -> exit 0; stderr/stdout never name foreign path
  -> no dirty preflight Error: for outside tree

# nested linked still cascades
consumer wt + deps/foo (git worktree add into consumer tree)
  -> Scan lists deps/foo (top-level and/or inner Worktrees)
  -> cascade removes deps/foo; own merge-back succeeds
```

## Preconditions

- Coverage backfill / mixed: foreign-isolation leaves are **GREEN OK** after P1
  (`scan_repo` base-path filter). Inner-worktree cascade leaf validates existing
  cascade success and documents ListWorktrees + inner field contract for P2;
  may already GREEN if FS walk surfaces `deps/foo` as a top-level worktree row
  without reading `Worktrees` — still a regression guard for collection source.
- Git required; Go required for inner-worktree leaf (consumer `go.mod`).
- Inherits monotree root helpers (`setupWrkWorktreeFromMain`, `runWrkFrom`,
  `initGitRepoOnMain`, worktree list asserts). Do **not** call `done-output/`
  helpers (sibling tree; not in inheritance chain).
- `SecondRepo` = abs path of foreign dirty tree (outside consumer).
- `DepsLinkedWtDir` / `DepPath` = nested linked cascade target + its dep main.

## Steps

- Grouping only: shared fixture helpers; leaves set `req.Args`.

```go
import (
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/gitops/git/git_isolated"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}

// initLocalMainRepo creates a fully-committed main repo at path without using the
// session seed CoW cache. Parallel leaves that both call setupWrkWorktreeFromMain
// can lose a race: isValidGitRepo sees .git before the seed's first commit, clone
// yields empty HEAD, and wrk worktree add fails with "invalid reference: HEAD".
func initLocalMainRepo(t *testing.T, path string) {
	t.Helper()
	mkdirAll(t, path)
	if err := git_isolated.Init(path, "main"); err != nil {
		t.Fatalf("git init %s: %v", path, err)
	}
	writeFile(t, filepath.Join(path, "README.md"), "# test\n")
	runGitIsolated(t, path, "add", "README.md")
	runGitIsolated(t, path, "commit", "-m", "init")
	// Hard-fail early if HEAD is unborn (never hand an empty repo to wrk).
	_ = gitOutputIsolated(t, path, "rev-parse", "HEAD")
}

// setupCascadeScanForeignIsolation builds:
//   - clean consumer main + wrk-managed linked worktree (no nested cascade targets)
//   - dirty independent main at WorkRoot/other/external/agent-pro (OUTSIDE consumerTop)
// Sets MainRepo, WtDir, WtBranch, SecondRepo (foreign abs), RepoDir=WtDir.
// Caller sets Args (--done or --done --dry-run).
func setupCascadeScanForeignIsolation(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)

	// Local init (not setupWrkWorktreeFromMain / seed CoW): sibling foreign leaves
	// run in parallel and must not race on fixtureSeedMainGoMod mid-build.
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	initLocalMainRepo(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/myrepo\n\ngo 1.21\n")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add go.mod")
	_ = gitOutputIsolated(t, mainRepo, "rev-parse", "HEAD")

	wtDir := runWrkFrom(t, req, mainRepo)
	req.MainRepo = compositionResolvePath(t, mainRepo)
	req.WtDir = compositionResolvePath(t, wtDir)
	req.WtBranch = branchName("main", wrkDate, 0)

	// Sibling layout mirrors pre-P1 leak fixture names (other/external/agent-pro)
	// but path is intentionally outside the consumer checkout tree.
	// Also local-init (not session seed) so parallel leaves never CoW half-built foreign.
	foreign := filepath.Join(req.WorkRoot, "other", "external", "agent-pro")
	initLocalMainRepo(t, foreign)
	writeFile(t, filepath.Join(foreign, "README"), "foreign main\n")
	runGitIsolated(t, foreign, "add", "README")
	runGitIsolated(t, foreign, "commit", "-m", "foreign initial")
	// Dirty after commit — would block --done if wrongly collected as cascade target.
	writeFile(t, filepath.Join(foreign, "dirty-foreign"), "uncommitted outside consumer")
	req.SecondRepo = compositionResolvePath(t, foreign)

	req.RepoDir = req.WtDir
}

// setupCascadeScanInnerWorktree builds consumer linked wt + manual deps/foo
// linked worktree of a separate dep main (same pattern as cascade-non-external-linked).
// Documents P2: cascade must still discover deps/foo via Scan (ListWorktrees +
// inner Worktrees and/or top-level worktree row under consumer).
func setupCascadeScanInnerWorktree(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)
	if _, err := exec.LookPath("go"); err != nil {
		t.Fatalf("go not found: %v", err)
	}

	mainRepo := filepath.Join(req.WorkRoot, "consumer")
	req.MainRepo = mainRepo
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "go.mod"), "module example.com/consumer\n\ngo 1.22\n")
	runGitIsolated(t, mainRepo, "add", "go.mod")
	runGitIsolated(t, mainRepo, "commit", "-m", "add consumer go.mod")

	wtDir := runWrkFrom(t, req, mainRepo)
	wtDir = compositionResolvePath(t, wtDir)
	req.WtDir = wtDir
	req.WtBranch = branchName("main", wrkDate, 0)

	depRepo := filepath.Join(req.WorkRoot, "foodep")
	req.DepPath = compositionResolvePath(t, depRepo)
	initGitRepoOnMain(t, depRepo)
	writeFile(t, filepath.Join(depRepo, "go.mod"), "module example.com/foodep\n\ngo 1.22\n")
	writeFile(t, filepath.Join(depRepo, "foo.go"), "package foo\n")
	runGitIsolated(t, depRepo, "add", "go.mod", "foo.go")
	runGitIsolated(t, depRepo, "commit", "-m", "add module")

	depsWtDir := filepath.Join(wtDir, "deps", "foo")
	req.DepsLinkedWtDir = compositionResolvePath(t, depsWtDir)
	depBranch := branchName("main", wrkDate, 0)
	runGitIsolated(t, depRepo, "worktree", "add", "-b", depBranch, depsWtDir)

	// Keep consumer porcelain clean so --done is not blocked by untracked deps/.
	writeFile(t, filepath.Join(wtDir, ".gitignore"), "/deps\n")
	runGitIsolated(t, wtDir, "add", ".gitignore")
	runGitIsolated(t, wtDir, "commit", "-m", "ignore deps worktrees")

	req.MainRepo = compositionResolvePath(t, mainRepo)
	req.RepoDir = wtDir
}

// assertNoForeignPathInOutput fails if stderr/stdout mention the foreign abs path
// or its distinctive basename (agent-pro). Used so dirty-outside cannot surface
// as cascade preflight Error: or cascade plan context.
func assertNoForeignPathInOutput(t *testing.T, req *Request, resp *Response) {
	t.Helper()
	if req.SecondRepo == "" {
		t.Fatal("SecondRepo (foreign path) must be set")
	}
	foreignBase := filepath.Base(req.SecondRepo)
	combined := resp.Stdout + "\n" + resp.Stderr
	if strings.Contains(combined, req.SecondRepo) {
		t.Fatalf("output must not mention foreign path %q; stdout=%q stderr=%q",
			req.SecondRepo, resp.Stdout, resp.Stderr)
	}
	// Basename alone can appear in unrelated strings; require path-ish context:
	// reject only when foreign base is paired with dirty/Error cascade language,
	// or when the path segment other/external/agent-pro appears.
	if strings.Contains(combined, "other/external/"+foreignBase) ||
		strings.Contains(combined, filepath.Join("other", "external", foreignBase)) {
		t.Fatalf("output must not mention foreign layout …/other/external/%s; stdout=%q stderr=%q",
			foreignBase, resp.Stdout, resp.Stderr)
	}
	// Hard fail if Error: mentions the foreign basename (dirty preflight leak).
	if strings.Contains(resp.Stderr, "Error:") && strings.Contains(resp.Stderr, foreignBase) {
		t.Fatalf("stderr Error: must not name foreign base %q; stderr=%q", foreignBase, resp.Stderr)
	}
}

var (
	_ = initLocalMainRepo
	_ = setupCascadeScanForeignIsolation
	_ = setupCascadeScanInnerWorktree
	_ = assertNoForeignPathInOutput
)
```
