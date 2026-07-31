# Scenario

**Feature**: bare `wrk --pr --title T --comment C` creates/attaches a GitHub PR via `gh`

```
# linked worktree + github.com origin + gh on PATH
linked wt (feature) + origin (github.com fetch URL + bare pushurl)
  -> wrk --pr --title T --comment C
  -> ensure remote head branch (push only if missing)
  -> gh pr list/create/comment
  -> stdout success tokens + PR URL; event command=pr

# refuse / validation
  main repo / exclusive modes / non-github / detached / missing gh / missing title|comment
  -> non-zero, clear stderr
```

## Preconditions

- Monotree root harness (`Request` / `Response` / `Run`, `runGitIsolated`, `v2StdoutTemplate`, …).
- Classic TDD: bare `--pr` is **not** implemented yet. Leaves document the target contract and must RED until implementer.
- L2 in-process (`req.InProcess = true`) with fake `gh` via `PathPrepend` (same pattern as create-ux fake `agent-run`).
- Hermetic origin: fetch URL is `https://github.com/acme/app.git` (passes `isGitHubRemoteURL`); `remote.origin.pushurl` points at a local bare repo so ensure-push is real.
- **P2 compose** (classic TDD under `compose/`): `--push --pr` always full branch push then PR; optional `--gen-commit-msg --commit` before push/pr. Bare `--pr` ensure-only semantics stay in `create-new/`.
- **P3 polish** (coverage leaves): color under `color/`; skill + bash-integration flag locks live in sibling trees (`cmd/wrk/tests/skill/`, `cmd/wrk/tests/bash-integration/`).
- **Not** covered here: dry-run for `--pr`, GHE, `--done`/`--merge-back` with PR.

## Steps

- Grouping/leaves call fixture helpers and set `req.Args` / `req.RepoDir` / fake-gh env.

## Context

- Base branch for PR = main checkout's **current** branch (fixture keeps `main`).
- Head = linked worktree current branch.
- New PR: `--comment` is initial body (`gh pr create --body`); no issue comment.
- Existing PR: title ignored; `--comment` is additive `gh pr comment` only.
- Existing open PR: title ignored (warning on stderr); still comment + URL.

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/gitops/git/git_isolated"
)

const (
	prDefaultTitle   = "Fix login"
	prDefaultComment = "please review"
	prDefaultURL     = "https://github.com/acme/app/pull/42"
	prGithubOrigin   = "https://github.com/acme/app.git"
	prFeatureBranch  = "feature-pr"
	prExistingTitle  = "Fix login"
	prExistingNumber = 42

	envFakeGhLog       = "FAKE_GH_LOG"
	envFakeGhListJSON  = "FAKE_GH_LIST_JSON"
	envFakeGhCreateURL = "FAKE_GH_CREATE_URL"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	prEnsureHelpersUsed()
	return nil
}

// --- fake gh (PathPrepend) ---

func installFakeGh(t *testing.T, req *Request) {
	t.Helper()
	binDir := filepath.Join(req.WorkRoot, "gh-bin")
	mkdirAll(t, binDir)
	req.PathPrepend = binDir

	logPath := filepath.Join(req.WorkRoot, "fake-gh.log")
	req.InterceptorLog = logPath
	// Truncate log for this leaf.
	writeFile(t, logPath, "")

	req.ExtraEnv = append(req.ExtraEnv,
		envFakeGhLog+"="+logPath,
		envFakeGhCreateURL+"="+prDefaultURL,
		// Default: no open PR. Leaves with existing PR override FAKE_GH_LIST_JSON.
		envFakeGhListJSON+"=[]",
	)

	body := `#!/bin/sh
log="${FAKE_GH_LOG:-}"
if [ -n "$log" ]; then
  {
    cmd_name=$(basename "$0")
    printf 'ARGC %s\n' "$(($# + 1))"
    for a in "$cmd_name" "$@"; do
      len=$(printf '%s' "$a" | wc -c | tr -d ' \t')
      printf 'LEN %s\n' "$len"
      printf '%s' "$a"
      printf '\n'
    done
    printf '---\n'
  } >> "$log"
fi

if [ "$1" = "pr" ] && [ "$2" = "list" ]; then
  printf '%s\n' "${FAKE_GH_LIST_JSON:-[]}"
  exit 0
fi
if [ "$1" = "pr" ] && [ "$2" = "create" ]; then
  printf '%s\n' "${FAKE_GH_CREATE_URL:-https://github.com/acme/app/pull/42}"
  exit 0
fi
if [ "$1" = "pr" ] && [ "$2" = "comment" ]; then
  exit 0
fi
printf 'fake-gh: unhandled argv:' >&2
printf ' %s' "$@" >&2
printf '\n' >&2
exit 1
`
	fake := filepath.Join(binDir, "gh")
	if err := os.WriteFile(fake, []byte(body), 0o755); err != nil {
		t.Fatalf("write fake gh: %v", err)
	}
}

// installPathWithoutGh sets PATH to a bin dir that has git but not gh so
// exec.LookPath("gh") fails. Used by refuse/missing-gh.
func installPathWithoutGh(t *testing.T, req *Request) {
	t.Helper()
	binDir := filepath.Join(req.WorkRoot, "no-gh-bin")
	mkdirAll(t, binDir)

	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Fatalf("LookPath git: %v", err)
	}
	link := filepath.Join(binDir, "git")
	if err := os.Symlink(gitPath, link); err != nil {
		t.Fatalf("symlink git: %v", err)
	}

	// No PathPrepend (would not add gh). Force PATH to only this bin.
	req.PathPrepend = ""
	req.ExtraEnv = append(req.ExtraEnv, "PATH="+binDir)
}

func setFakeGhExistingPR(t *testing.T, req *Request, title, url string, number int) {
	t.Helper()
	if title == "" {
		title = prExistingTitle
	}
	if url == "" {
		url = prDefaultURL
	}
	if number == 0 {
		number = prExistingNumber
	}
	// Compact JSON array with one open PR (fields product may parse).
	js := fmt.Sprintf(`[{"number":%d,"title":%q,"url":%q}]`, number, title, url)
	// Replace or append FAKE_GH_LIST_JSON.
	replaced := false
	for i, e := range req.ExtraEnv {
		if strings.HasPrefix(e, envFakeGhListJSON+"=") {
			req.ExtraEnv[i] = envFakeGhListJSON + "=" + js
			replaced = true
			break
		}
	}
	if !replaced {
		req.ExtraEnv = append(req.ExtraEnv, envFakeGhListJSON+"="+js)
	}
}

// --- git fixtures (github origin + pushurl + linked wt) ---

func setupPrBareOrigin(t *testing.T, workRoot, name string) string {
	t.Helper()
	bare := filepath.Join(workRoot, name+".git")
	runGitIsolated(t, workRoot, "-c", "init.templateDir=", "init", "--bare", "-b", "main", bare)
	return bare
}

// configureGithubOriginPushURL sets origin fetch URL to github.com shape and
// pushurl to local bare so isGitHubRemoteURL passes and git push is hermetic.
func configureGithubOriginPushURL(t *testing.T, repo, bare string) {
	t.Helper()
	runGitIsolated(t, repo, "remote", "add", "origin", prGithubOrigin)
	runGitIsolated(t, repo, "config", "remote.origin.pushurl", bare)
}

// setupPrMainWithGithubOrigin: main checkout, github-shaped origin, pushurl bare,
// main pushed to origin. Local main tip == origin/main.
func setupPrMainWithGithubOrigin(t *testing.T, req *Request) {
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

	bare := setupPrBareOrigin(t, req.WorkRoot, "origin")
	configureGithubOriginPushURL(t, mainRepo, bare)
	runGitIsolated(t, mainRepo, "push", "-u", "origin", "main")
	req.OriginBare = bare
}

// setupPrLinkedFeature: main+github origin; linked worktree on feature branch
// with one commit. By default the feature branch is NOT on origin (ensure-push path).
func setupPrLinkedFeature(t *testing.T, req *Request) {
	t.Helper()
	setupPrMainWithGithubOrigin(t, req)

	feature := prFeatureBranch
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

// setupPrLinkedFeatureRemoteExists: like setupPrLinkedFeature, then push feature
// to origin so remote head already exists (no ensure-push expected).
func setupPrLinkedFeatureRemoteExists(t *testing.T, req *Request) {
	t.Helper()
	setupPrLinkedFeature(t, req)
	runGitIsolated(t, req.WtDir, "push", "-u", "origin", req.WtBranch)
}

// setupPrLinkedFeatureRemoteExistsLocalAhead: remote has feature tip; local is
// one commit ahead. --pr must NOT push the tip (only ensure when missing).
func setupPrLinkedFeatureRemoteExistsLocalAhead(t *testing.T, req *Request) {
	t.Helper()
	setupPrLinkedFeatureRemoteExists(t, req)
	// Snapshot remote tip before local ahead commit.
	remoteSHA := revParseRef(t, req.OriginBare, "refs/heads/"+req.WtBranch)
	writeFile(t, filepath.Join(req.WorkRoot, "origin-feature-before"), remoteSHA+"\n")

	writeFile(t, filepath.Join(req.WtDir, "ahead.txt"), "local only\n")
	runGitIsolated(t, req.WtDir, "add", "ahead.txt")
	runGitIsolated(t, req.WtDir, "commit", "-m", "local ahead of origin")
}

// setupPrNonGithubOrigin: linked feature with bare-path origin only (not github.com).
func setupPrNonGithubOrigin(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)

	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "README.md"), "# main\n")
	runGitIsolated(t, mainRepo, "add", "README.md")
	runGitIsolated(t, mainRepo, "commit", "-m", "init")
	mainRepo = compositionResolvePath(t, mainRepo)
	req.MainRepo = mainRepo

	bare := setupPrBareOrigin(t, req.WorkRoot, "origin")
	runGitIsolated(t, mainRepo, "remote", "add", "origin", bare)
	runGitIsolated(t, mainRepo, "push", "-u", "origin", "main")
	req.OriginBare = bare

	feature := prFeatureBranch
	linked := filepath.Join(req.WorkRoot, "linked-wt")
	runGitIsolated(t, mainRepo, "worktree", "add", "-b", feature, linked)
	linked = compositionResolvePath(t, linked)
	writeFile(t, filepath.Join(linked, "feature.txt"), "from linked wt\n")
	runGitIsolated(t, linked, "add", "feature.txt")
	runGitIsolated(t, linked, "commit", "-m", "feature work")

	req.WtDir = linked
	req.WtBranch = feature
	req.RepoDir = linked
}

// setupPrDetachedLinked: linked worktree then detach HEAD.
func setupPrDetachedLinked(t *testing.T, req *Request) {
	t.Helper()
	setupPrLinkedFeatureRemoteExists(t, req)
	runGitIsolated(t, req.WtDir, "checkout", "--detach")
	req.HashToken = strings.TrimSpace(gitOutputIsolated(t, req.WtDir, "rev-parse", "--short=7", "HEAD"))
}

// prArgs builds the standard successful --pr argv.
func prArgs(title, comment string) []string {
	return []string{"--pr", "--title", title, "--comment", comment}
}

func prDefaultArgs() []string {
	return prArgs(prDefaultTitle, prDefaultComment)
}

// --- stdout shapes ---

func prPushConfirmLine(branch string) string {
	if branch == "" {
		branch = prFeatureBranch
	}
	return fmt.Sprintf("pushed %s → origin/%s\n", branch, branch)
}

func prCreatedStdout(title, url string) string {
	if title == "" {
		title = prDefaultTitle
	}
	if url == "" {
		url = prDefaultURL
	}
	return fmt.Sprintf("PR created\ntitle set: %s\nbody set\n%s\n", title, url)
}

func prCreatedWithPushStdout(branch, title, url string) string {
	return prPushConfirmLine(branch) + prCreatedStdout(title, url)
}

func prExistingStdout(url string) string {
	if url == "" {
		url = prDefaultURL
	}
	return fmt.Sprintf("comment added\n%s\n", url)
}

func prExistingTitleWarning(existingTitle string) string {
	if existingTitle == "" {
		existingTitle = prExistingTitle
	}
	return fmt.Sprintf("warning: title ignored (PR already exists); existing title: %s\n", existingTitle)
}

// --- fake gh log parse ---

type prGhInvoc struct {
	Args []string
}

func parseFakeGhLog(t *testing.T, logPath string) []prGhInvoc {
	t.Helper()
	data, err := os.ReadFile(logPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read fake gh log: %v", err)
	}
	lines := strings.Split(string(data), "\n")
	var invocs []prGhInvoc
	var cur prGhInvoc
	i := 0
	for i < len(lines) {
		line := lines[i]
		if line == "---" || line == "" {
			if len(cur.Args) > 0 {
				invocs = append(invocs, cur)
				cur = prGhInvoc{}
			}
			i++
			continue
		}
		if strings.HasPrefix(line, "ARGC ") {
			// next pairs of LEN + body
			i++
			for i < len(lines) && strings.HasPrefix(lines[i], "LEN ") {
				i++ // skip LEN
				if i >= len(lines) {
					break
				}
				cur.Args = append(cur.Args, lines[i])
				i++
			}
			continue
		}
		i++
	}
	if len(cur.Args) > 0 {
		invocs = append(invocs, cur)
	}
	return invocs
}

func ghLogPath(req *Request) string {
	if req.InterceptorLog != "" {
		return req.InterceptorLog
	}
	return filepath.Join(req.WorkRoot, "fake-gh.log")
}

func assertGhSubcmdCalled(t *testing.T, invocs []prGhInvoc, sub string) prGhInvoc {
	t.Helper()
	for _, inv := range invocs {
		// args: gh pr <sub> ...
		if len(inv.Args) >= 3 && inv.Args[0] == "gh" && inv.Args[1] == "pr" && inv.Args[2] == sub {
			return inv
		}
	}
	t.Fatalf("expected gh pr %s invocation; got %#v", sub, invocs)
	return prGhInvoc{}
}

func assertGhSubcmdNotCalled(t *testing.T, invocs []prGhInvoc, sub string) {
	t.Helper()
	for _, inv := range invocs {
		if len(inv.Args) >= 3 && inv.Args[0] == "gh" && inv.Args[1] == "pr" && inv.Args[2] == sub {
			t.Fatalf("expected no gh pr %s; got %#v", sub, inv)
		}
	}
}

func assertGhArgContains(t *testing.T, inv prGhInvoc, want string) {
	t.Helper()
	for _, a := range inv.Args {
		if a == want {
			return
		}
	}
	t.Fatalf("gh invoc args %v should contain %q", inv.Args, want)
}

// --- origin helpers (reuse push-style) ---

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

// --- events.jsonl ---

type prWrkEvent struct {
	TS       string   `json:"ts"`
	Command  string   `json:"command"`
	WorkDir  string   `json:"work_dir"`
	MainRepo string   `json:"main_repo"`
	Args     []string `json:"args"`
	ExitCode int      `json:"exit_code"`
}

func prEventsPath(wrkHome string) string {
	return filepath.Join(wrkHome, "events.jsonl")
}

func readPrEvents(t *testing.T, wrkHome string) []prWrkEvent {
	t.Helper()
	data, err := os.ReadFile(prEventsPath(wrkHome))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read events.jsonl: %v", err)
	}
	var events []prWrkEvent
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var ev prWrkEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			t.Fatalf("parse event line %q: %v", line, err)
		}
		events = append(events, ev)
	}
	return events
}

func prEnsureHelpersUsed() {
	_ = installFakeGh
	_ = installPathWithoutGh
	_ = setFakeGhExistingPR
	_ = setupPrBareOrigin
	_ = configureGithubOriginPushURL
	_ = setupPrMainWithGithubOrigin
	_ = setupPrLinkedFeature
	_ = setupPrLinkedFeatureRemoteExists
	_ = setupPrLinkedFeatureRemoteExistsLocalAhead
	_ = setupPrNonGithubOrigin
	_ = setupPrDetachedLinked
	_ = prArgs
	_ = prDefaultArgs
	_ = prPushConfirmLine
	_ = prCreatedStdout
	_ = prCreatedWithPushStdout
	_ = prExistingStdout
	_ = prExistingTitleWarning
	_ = parseFakeGhLog
	_ = ghLogPath
	_ = assertGhSubcmdCalled
	_ = assertGhSubcmdNotCalled
	_ = assertGhArgContains
	_ = revParseRef
	_ = assertOriginBranchEqualsLocal
	_ = originBranchExists
	_ = readPrEvents
}
```
