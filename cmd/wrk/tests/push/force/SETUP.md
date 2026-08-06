# Scenario

**Feature**: `-f` / `--force` with `--push` force-with-lease pushes the current checkout branch

```
# force is only valid when --push is also set
wrk -f / wrk --force
  -> non-zero
  -> wrk: -f/--force is only valid with --push

# with --push: branch push uses git push --force-with-lease
myrepo (main or linked wt) + origin
  -> wrk --push -f | --push --force
  -> git push --force-with-lease <remote> <branch>
  -> stdout: pushed <branch> → <remote>/<remoteBranch>   # no "force-pushed"
  -> origin refs/heads/<branch> == local HEAD (incl. non-FF when lease ok)

# dry-run
  -> wrk --push -f --dry-run
  -> would: git push --force-with-lease origin main
  -> no remote mutation

# flag order free: -f --push ≡ --push -f
```

## Preconditions

- Parent `push/SETUP.md` helpers: `setupPushMainWithOrigin`, `setupPushFromLinkedWorktree`,
  `pushConfirmLine`, `wouldPushBranchLine`, `assertOriginBranchEqualsLocal`, `revParseRef`, …
- Monotree root harness (`Request` / `Response` / `Run`, `v2StdoutTemplate`, `req.InProcess` L2).
- Classic TDD for **force**: product does not yet accept `-f`/`--force` on push paths
  (today: unrecognized flag). Leaves encode the target contract and must RED until GREEN.
- **Locks**: force → `--force-with-lease` (not bare `--force`); confirm line stays
  `pushed … → …`; force without `--push` hard-errors with the exact message above.
- Tags stay non-force when also pushed (out of scope here — branch-only bare `--push`).
- Full done/merge-back/tag-next/pr force matrix is out of scope (optional compose smoke max).

## Steps

- Grouping: leaves call fixture helpers and set `req.Args` / `req.RepoDir` with `req.InProcess = true`.

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
	forceEnsureHelpersUsed()
	return nil
}

// forceWithoutPushErr is the locked hard-error when -f/--force is set without --push.
const forceWithoutPushErr = "wrk: -f/--force is only valid with --push"

// wouldForcePushBranchLine is the dry-run plan line for force-with-lease branch push.
func wouldForcePushBranchLine(remote, branch string) string {
	if remote == "" {
		remote = "origin"
	}
	if branch == "" {
		branch = "main"
	}
	return fmt.Sprintf("would: git push --force-with-lease %s %s\n", remote, branch)
}

// setupPushDivergedMainWithOrigin builds main + bare origin where local main and
// origin/main have diverged (shared ancestor, different tips). Plain
// `git push` / `wrk --push` must fail non-FF; `git push --force-with-lease`
// succeeds when the lease matches origin's tip.
//
// After return: local HEAD is the local-only tip; origin/main is the remote-only tip.
// Snapshots origin tip at {WorkRoot}/origin-main-before for Assert.
func setupPushDivergedMainWithOrigin(t *testing.T, req *Request) {
	t.Helper()
	setupPushMainWithOrigin(t, req)

	base := revParseHEAD(t, req.MainRepo)

	// Publish a remote-only tip from local, then rewind local and commit differently.
	writeFile(t, filepath.Join(req.MainRepo, "remote-only.txt"), "on origin only\n")
	runGitIsolated(t, req.MainRepo, "add", "remote-only.txt")
	runGitIsolated(t, req.MainRepo, "commit", "-m", "remote-only tip")
	runGitIsolated(t, req.MainRepo, "push", "origin", "main")

	runGitIsolated(t, req.MainRepo, "reset", "--hard", base)
	writeFile(t, filepath.Join(req.MainRepo, "local-only.txt"), "local tip only\n")
	runGitIsolated(t, req.MainRepo, "add", "local-only.txt")
	runGitIsolated(t, req.MainRepo, "commit", "-m", "local-only tip")

	originSHA := revParseRef(t, req.OriginBare, "refs/heads/main")
	localSHA := revParseHEAD(t, req.MainRepo)
	if originSHA == localSHA {
		t.Fatal("fixture expected diverged tips: origin/main == local HEAD")
	}
	writeFile(t, filepath.Join(req.WorkRoot, "origin-main-before"), originSHA+"\n")
}

func readOriginMainBefore(t *testing.T, req *Request) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(req.WorkRoot, "origin-main-before"))
	if err != nil {
		t.Fatalf("read origin-main-before: %v", err)
	}
	return strings.TrimSpace(string(b))
}

func forceEnsureHelpersUsed() {
	_ = wouldForcePushBranchLine
	_ = setupPushDivergedMainWithOrigin
	_ = readOriginMainBefore
	_ = forceWithoutPushErr
}
```
