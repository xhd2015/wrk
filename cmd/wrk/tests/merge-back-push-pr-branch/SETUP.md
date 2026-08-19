# Scenario

**Feature**: `[--merge-back|--done] --push` also lease-updates `origin/<worktree-branch>` when that ref already exists

```
# standing PR branch on origin; land + --push must not leave it on pre-rebase commits
myrepo (origin/main + origin/<WtBranch>) + wt
  -> wrk --merge-back|--done -y --push
  -> land, then pushed main → origin/main
  -> when origin/<WtBranch> tip is unchanged and already in local: lease-push post-land tip
  -> when tip is not in local: warning; origin/<WtBranch> left alone; exit 0
```

## Preconditions

- Git available; monotree root helpers (`runWrkFrom`, `commitAheadOnWorktree`, `revParseHEAD`, `v2StdoutTemplate`, …).
- Distinct from `done-push/` (main-only, no `origin/<WtBranch>`) and `merge-back-pipeline/` (same).

## Steps

- Grouping only: leaves call fixture helpers and set `req.Args`.

```go
import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}

func setupPRBranchBareOrigin(t *testing.T, workRoot, name string) string {
	t.Helper()
	bare := filepath.Join(workRoot, name+".git")
	runGitIsolated(t, workRoot, "-c", "init.templateDir=", "init", "--bare", "-b", "main", bare)
	return bare
}

func revParseRefPR(t *testing.T, repo, ref string) string {
	t.Helper()
	return strings.TrimSpace(gitOutputIsolated(t, repo, "rev-parse", ref))
}

func pushWorktreeBranchToOrigin(t *testing.T, wtDir, originBare, branch string) {
	t.Helper()
	runGitIsolated(t, wtDir, "push", "origin", "HEAD:refs/heads/"+branch)
	got := revParseRefPR(t, originBare, "refs/heads/"+branch)
	want := revParseHEAD(t, wtDir)
	if got != want {
		t.Fatalf("origin/%s %s != worktree HEAD %s", branch, got, want)
	}
}

// setupAheadWithOriginBranch: main + origin/main + wrk wt ahead + origin/<WtBranch> at wt HEAD.
func setupAheadWithOriginBranch(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)

	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	cloneMainGoModFromSeed(t, mainRepo)
	mainRepo = compositionResolvePath(t, mainRepo)
	req.MainRepo = mainRepo

	bare := setupPRBranchBareOrigin(t, req.WorkRoot, "origin")
	runGitIsolated(t, mainRepo, "remote", "add", "origin", bare)
	runGitIsolated(t, mainRepo, "push", "-u", "origin", "main")
	req.OriginBare = bare

	wtDir := runWrkFrom(t, req, mainRepo)
	wtDir = compositionResolvePath(t, wtDir)
	req.WtDir = wtDir
	req.WtBranch = branchName("main", wrkDate, 0)

	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead of main")
	pushWorktreeBranchToOrigin(t, wtDir, bare, req.WtBranch)
	req.RepoDir = wtDir
}

// setupDivergedWithOriginBranch: same as ahead, then an extra commit on main so land rebases.
func setupDivergedWithOriginBranch(t *testing.T, req *Request) {
	t.Helper()
	setupAheadWithOriginBranch(t, req)
	writeFile(t, filepath.Join(req.MainRepo, "main-extra"), "from main\n")
	runGitIsolated(t, req.MainRepo, "add", "main-extra")
	runGitIsolated(t, req.MainRepo, "commit", "-m", "main extra")
}

// setupNotIncludedOriginBranch: origin/<WtBranch> has a commit that is not in the worktree.
func setupNotIncludedOriginBranch(t *testing.T, req *Request) {
	t.Helper()
	setupAheadWithOriginBranch(t, req)
	tmp := filepath.Join(req.WorkRoot, "tmp-remote-only")
	runGitIsolated(t, req.WorkRoot, "clone", req.OriginBare, tmp)
	runGitIsolated(t, tmp, "checkout", req.WtBranch)
	writeFile(t, filepath.Join(tmp, "remote-only"), "from remote\n")
	runGitIsolated(t, tmp, "add", "remote-only")
	runGitIsolated(t, tmp, "commit", "-m", "remote-only")
	runGitIsolated(t, tmp, "push", "origin", req.WtBranch)
	writeFile(t, filepath.Join(req.WorkRoot, "origin-branch-tip"), revParseRefPR(t, req.OriginBare, "refs/heads/"+req.WtBranch)+"\n")
}

func loadSavedOriginBranchTip(t *testing.T, req *Request) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(req.WorkRoot, "origin-branch-tip"))
	if err != nil {
		t.Fatalf("read origin-branch-tip: %v", err)
	}
	return strings.TrimSpace(string(data))
}

func primaryThenTwoPushStdout(primaryMsg, branch string) string {
	primary := strings.TrimSuffix(primaryMsg, "\n") + "\n"
	return primary + "\n" +
		"pushed main → origin/main\n" +
		fmt.Sprintf("pushed %s → origin/%s\n", branch, branch)
}

func assertOriginBranchEquals(t *testing.T, originBare, branch, wantSHA string) {
	t.Helper()
	got := revParseRefPR(t, originBare, "refs/heads/"+branch)
	if got != wantSHA {
		t.Fatalf("origin/%s %s != want %s", branch, got, wantSHA)
	}
}

var (
	_ = setupAheadWithOriginBranch
	_ = setupDivergedWithOriginBranch
	_ = setupNotIncludedOriginBranch
	_ = loadSavedOriginBranchTip
	_ = primaryThenTwoPushStdout
	_ = assertOriginBranchEquals
)
```
