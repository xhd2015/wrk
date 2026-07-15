# Scenario

**Feature**: `wrk --done --push` pushes main branch tip to origin after successful done

```
# primary --done succeeds (not aborted) then runPushMain(main, tags=[])
linked wt (ahead) + bare origin on main
  -> wrk --done -y --push
  -> merge-back --rm (message on stdout)
  -> blank line
  -> pushed main → origin/main
  -> origin/main == post-merge main HEAD; worktree gone
```

## Preconditions

- Git available; monotree root helpers (`setupWrkWorktreeFromMain`, `commitAheadOnWorktree`, `revParseHEAD`, `v2StdoutTemplate`, …).
- **Branch-only** push after done (empty tags). No `--tag-next` here (tag+branch push is `done-pipeline/tag-next-push`); merge-back push is `merge-back-pipeline/`; multi-stage dry-run is `done-pipeline/dry-run/`.
- Locked: `runPushMain` runs after successful `--done` when `--push` is set (GREEN).

## Steps

- Grouping only: leaves call `setupDonePushWithOrigin` / no-remote fixture and set `req.Args`.

```go
import (
	"fmt"
	"path/filepath"
	"strings"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}

func setupDonePushBareOrigin(t *testing.T, workRoot, name string) string {
	t.Helper()
	bare := filepath.Join(workRoot, name+".git")
	runGitIsolated(t, workRoot, "-c", "init.templateDir=", "init", "--bare", "-b", "main", bare)
	return bare
}

// setupDonePushWithOrigin: seed main + bare origin (upstream set) + wrk wt ahead.
// RepoDir is the linked worktree. Caller sets Args.
func setupDonePushWithOrigin(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)

	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	cloneRepoFromSeed(t, fixtureSeedMainGoMod, buildSeedMainGoMod, mainRepo)
	mainRepo = compositionResolvePath(t, mainRepo)
	req.MainRepo = mainRepo

	bare := setupDonePushBareOrigin(t, req.WorkRoot, "origin")
	runGitIsolated(t, mainRepo, "remote", "add", "origin", bare)
	runGitIsolated(t, mainRepo, "push", "-u", "origin", "main")
	req.OriginBare = bare

	wtDir := runWrkFrom(t, req, mainRepo)
	wtDir = compositionResolvePath(t, wtDir)
	req.WtDir = wtDir
	req.WtBranch = branchName("main", wrkDate, 0)

	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead of main")
	req.RepoDir = wtDir
}

// setupDonePushNoRemote: linked wt ahead of main, no origin remote configured.
func setupDonePushNoRemote(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)
	mainRepo, wtDir, _ := setupWrkWorktreeFromMain(t, req)
	mainRepo = compositionResolvePath(t, mainRepo)
	req.MainRepo = mainRepo
	wtDir = compositionResolvePath(t, wtDir)
	req.WtDir = wtDir
	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead of main")
	req.RepoDir = wtDir
}

// primaryThenPushStdout joins primary MergeBack message and push line with a blank line.
func primaryThenPushStdout(primaryMsg, pushLine string) string {
	primary := strings.TrimSuffix(primaryMsg, "\n") + "\n"
	push := strings.TrimSuffix(pushLine, "\n") + "\n"
	return primary + "\n" + push
}

// donePushConfirmLine is the stable human confirmation for branch-only push (empty tags).
func donePushConfirmLine() string {
	return "pushed main → origin/main\n"
}

func revParseRef(t *testing.T, repo, ref string) string {
	t.Helper()
	return strings.TrimSpace(gitOutputIsolated(t, repo, "rev-parse", ref))
}

func assertOriginMainEqualsLocalMain(t *testing.T, mainRepo, originBare string) {
	t.Helper()
	mainSHA := revParseHEAD(t, mainRepo)
	originSHA := revParseRef(t, originBare, "refs/heads/main")
	if originSHA != mainSHA {
		t.Fatalf("origin/main %s != local main HEAD %s", originSHA, mainSHA)
	}
}

// keep helpers referenced for inheritance compilation
var (
	_ = setupDonePushWithOrigin
	_ = setupDonePushNoRemote
	_ = primaryThenPushStdout
	_ = donePushConfirmLine
	_ = assertOriginMainEqualsLocalMain
	_ = fmt.Sprintf
)
```
