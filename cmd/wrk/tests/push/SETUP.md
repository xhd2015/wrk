# Scenario

**Feature**: bare `wrk --push` pushes the current checkout branch to its upstream/origin

```
# option R: push current checkout branch (main or linked worktree branch)
git checkout (main or linked wt) + bare origin
  -> wrk --push
  -> git push <remote> <branch>
  -> stdout: pushed <branch> → <remote>/<remoteBranch>
  -> origin refs/heads/<branch> == local HEAD

# dry-run
  -> wrk --push --dry-run
  -> would: git push origin <branch>
  -> no remote mutation

# force (see force/): -f/--force only with --push → git push --force-with-lease
  -> wrk --push -f [--dry-run]
  -> would: git push --force-with-lease origin <branch>  (dry-run)
  -> confirm stays: pushed <branch> → origin/<branch>
  -> without --push: wrk: -f/--force is only valid with --push

# errors
  -> no origin/upstream -> non-zero, clear stderr
  -> --push --json alone -> still invalid (json only with --tag-next)
  -> --push --list -> mutually exclusive
```


## Preconditions

- Monotree root harness (`Request` / `Response` / `Run`, `runGitIsolated`, `v2StdoutTemplate`, …).
- Classic TDD: bare `--push` is **not** implemented yet (currently
  `wrk: --push is only valid with --tag-next`). Leaves document the target contract and must RED.
- Reuses the same `runPushMain` semantics as `done-push/` / done-pipeline (upstream prefer, else `origin` + branch name; confirm line; dry-run `would:` lines).
- **Not** covered here: `--done/--merge-back --push`, `--tag-next --push` (see `done-push/`, `tag-next/push/`).

## Steps

- Grouping: leaves call fixture helpers and set `req.Args` / `req.RepoDir`.

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/gitops/git/git_isolated"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	pushEnsureHelpersUsed()
	return nil
}

func setupPushBareOrigin(t *testing.T, workRoot, name string) string {
	t.Helper()
	bare := filepath.Join(workRoot, name+".git")
	runGitIsolated(t, workRoot, "-c", "init.templateDir=", "init", "--bare", "-b", "main", bare)
	return bare
}

// setupPushMainWithOrigin: main checkout with origin remote; upstream set on main.
// Local main tip equals origin/main after setup. Caller sets Args.
func setupPushMainWithOrigin(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)

	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "README.md"), "# main\n")
	runGitIsolated(t, mainRepo, "add", "README.md")
	runGitIsolated(t, mainRepo, "commit", "-m", "init")

	mainRepo = compositionResolvePath(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.WtBranch = "main"

	bare := setupPushBareOrigin(t, req.WorkRoot, "origin")
	runGitIsolated(t, mainRepo, "remote", "add", "origin", bare)
	runGitIsolated(t, mainRepo, "push", "-u", "origin", "main")
	req.OriginBare = bare
}

// setupPushNoRemote: main checkout, no remotes configured.
func setupPushNoRemote(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "README.md"), "# main\n")
	runGitIsolated(t, mainRepo, "add", "README.md")
	runGitIsolated(t, mainRepo, "commit", "-m", "init")
	mainRepo = compositionResolvePath(t, mainRepo)
	req.MainRepo = mainRepo
	req.RepoDir = mainRepo
	req.WtBranch = "main"
}

// setupPushFromLinkedWorktree: main + origin on main; linked worktree on feature
// branch with one commit not on origin. Option R: wrk --push from the worktree
// must push the worktree branch tip.
func setupPushFromLinkedWorktree(t *testing.T, req *Request) {
	t.Helper()
	setupPushMainWithOrigin(t, req)

	feature := "feature-push"
	linked := filepath.Join(req.WorkRoot, "linked-wt")
	runGitIsolated(t, req.MainRepo, "worktree", "add", "-b", feature, linked)
	linked = compositionResolvePath(t, linked)
	writeFile(t, filepath.Join(linked, "feature.txt"), "from linked wt\n")
	runGitIsolated(t, linked, "add", "feature.txt")
	runGitIsolated(t, linked, "commit", "-m", "feature work")

	req.WtDir = linked
	req.WtBranch = feature
	req.RepoDir = linked
}

// pushConfirmLine is the stable human confirmation for branch push (no tags).
func pushConfirmLine(branch string) string {
	if branch == "" {
		branch = "main"
	}
	return fmt.Sprintf("pushed %s → origin/%s\n", branch, branch)
}

func wouldPushBranchLine(remote, branch string) string {
	if remote == "" {
		remote = "origin"
	}
	if branch == "" {
		branch = "main"
	}
	return fmt.Sprintf("would: git push %s %s\n", remote, branch)
}

func revParseRef(t *testing.T, repo, ref string) string {
	t.Helper()
	return strings.TrimSpace(gitOutputIsolated(t, repo, "rev-parse", ref))
}

func assertOriginBranchEqualsLocal(t *testing.T, localRepo, originBare, branch string) {
	t.Helper()
	localSHA := revParseHEAD(t, localRepo)
	originSHA := revParseRef(t, originBare, "refs/heads/"+branch)
	if originSHA != localSHA {
		t.Fatalf("origin/%s %s != local HEAD %s (repo=%s)", branch, originSHA, localSHA, localRepo)
	}
}

func originBranchExists(t *testing.T, originBare, branch string) bool {
	t.Helper()
	err := git_isolated.Command(originBare, "rev-parse", "--verify", "refs/heads/"+branch).Run()
	return err == nil
}

// --- events.jsonl (bare push primary command) ---

type pushWrkEvent struct {
	TS       string   `json:"ts"`
	Command  string   `json:"command"`
	WorkDir  string   `json:"work_dir"`
	MainRepo string   `json:"main_repo"`
	Args     []string `json:"args"`
	ExitCode int      `json:"exit_code"`
}

func pushEventsPath(wrkHome string) string {
	return filepath.Join(wrkHome, "events.jsonl")
}

func readPushEvents(t *testing.T, wrkHome string) []pushWrkEvent {
	t.Helper()
	data, err := os.ReadFile(pushEventsPath(wrkHome))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read events.jsonl: %v", err)
	}
	var events []pushWrkEvent
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var ev pushWrkEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			t.Fatalf("parse event line %q: %v", line, err)
		}
		events = append(events, ev)
	}
	return events
}

func pushEnsureHelpersUsed() {
	_ = setupPushMainWithOrigin
	_ = setupPushNoRemote
	_ = setupPushFromLinkedWorktree
	_ = pushConfirmLine
	_ = wouldPushBranchLine
	_ = assertOriginBranchEqualsLocal
	_ = originBranchExists
	_ = readPushEvents
	_ = revParseRef
}
```
