# Scenario

**Feature**: multi-mode `wrk --pr` — show, comment-only, create/attach, push-existing, status

```
# show mode (P1)
linked wt (feature) + origin (github.com) + gh on PATH
  -> wrk --pr
  -> gh pr list --head <branch> --state open
  -> open PR: stdout = URL only; no PR: empty stdout; exit 0
  -> no ensure-push / create / comment

# comment-only (P2)
linked wt + origin + gh
  -> wrk --pr --comment C   # no --title
  -> list open PR for head; never create; never ensure-push
  -> open PR: gh pr comment; stdout comment added + URL (no title-ignored warning)
  -> no open PR: non-zero; stderr mentions no open pull request

# push-existing (P3)
linked wt + origin + gh; --push + --pr; no --title
  -> wrk --pr --push  [optional --comment C]
  -> list open PR for head FIRST; if none: non-zero, origin tip UNCHANGED
  -> open PR: full tip push then (optional comment) + URL
  -> never gh pr create; create compose still needs --title + --comment

# status (P4 classic TDD — RED until implementer)
linked wt + origin + gh; --pr --status (no title/comment/push)
  -> list open PR for head; if open: gh pr view N --json …
  -> open PR: stdout URL + State/Title/Checks/Reviews rollup; exit 0 even if checks=failure
  -> no open PR: exit 0; stderr warning: no open …; stdout empty
  -> --pr --status + --title/--comment/--push: non-zero invalid combination
  -> flag order free: --status --pr same; never push/create/comment

# create / attach + create compose (GREEN under create-new/ + existing-pr/ + compose/)
linked wt + origin (github.com fetch URL + bare pushurl)
  -> wrk --pr --title T --comment C
  -> ensure remote head branch (push only if missing)
  -> gh pr list/create/comment
  -> stdout success tokens + PR URL; event command=pr
  -> wrk --push --pr --title T --comment C: full push then create/attach

# refuse / incomplete create
  main repo / exclusive modes / non-github / detached / missing gh
  -> non-zero, clear stderr
  --pr --title T (no --comment) / empty title|comment when both flags present
  -> non-zero (incomplete create; not "title always required with --comment alone")
```

## Preconditions

- Monotree root harness (`Request` / `Response` / `Run`, `runGitIsolated`, `v2StdoutTemplate`, …).
- **P1 show** (`show/`): bare `wrk --pr` URL-only or empty; stays GREEN.
- **P2 comment-only** (`comment-only/`): `--pr --comment C` without title; stays GREEN.
- **P3 push-existing** (`push-existing/`): `--pr --push` without title; open PR required → full tip push → URL.
- Classic TDD **P4**: `wrk --pr --status` **PR status** is **not** implemented yet (`status && prFlag` still mutual-exclusion; no `runPRStatus`). Leaves under `status/` document the target contract and must **RED** until implementer. Show, comment-only, create/attach, push-existing, and create compose stay **GREEN**. Global bare `wrk --status` (git worktree status) is out of this tree and must stay GREEN.
- L2 in-process (`req.InProcess = true`) with fake `gh` via `PathPrepend` (same pattern as create-ux fake `agent-run`).
- Hermetic origin: fetch URL is `https://github.com/acme/app.git` (passes `isGitHubRemoteURL`); `remote.origin.pushurl` points at a local bare repo so ensure-push / full push is real.
- **Create compose** under `compose/`: `--push --pr --title T --comment C` always full branch push then PR; optional `--gen-commit-msg --commit` before push/pr. Bare `--pr` ensure-only semantics stay in `create-new/` (create mode); show, comment-only, and status never push.
- **Later**: color / skill / bash-integration polish (P5).
- **Not** covered here: dry-run for `--pr`, GHE, `--done`/`--merge-back` with PR, full `gh pr checks` table (status uses compact Checks rollup only).

## Steps

- Grouping/leaves call fixture helpers and set `req.Args` / `req.RepoDir` / fake-gh env.

## Context

- Base branch for PR = main checkout's **current** branch (fixture keeps `main`).
- Head = linked worktree current branch.
- **Show** (`--pr` alone): list open PR for head; print URL or empty; no push/create/comment.
- **Comment-only** (`--pr --comment C`, no `--title`): attach comment to existing open PR only; error if none; no push/create; no title-ignored warning.
- **Push-existing** (`--pr --push` without `--title`): open PR required; list before any push; full tip push (not ensure-only); optional `--comment` after push; never create; flag order free (`--push --pr` same).
- **Status** (`--pr --status`, no title/comment/push): open-only list; when open, `gh pr view` JSON for state/title/checks/reviews; read-only; exit 0 for successful query even when Checks=failure; flag order free (`--status --pr` same). Mutex carve-out: `status && prFlag` allowed when neither other exclusive modes apply; resolveCommand stays `"pr"`. Invalid: status + title/comment/push.
- **Create compose** (`--push --pr --title T --comment C`): full push then create/attach (unchanged by P4).
- New PR: `--comment` is initial body (`gh pr create --body`); no issue comment.
- Existing PR (create/attach argv): title ignored; `--comment` is additive `gh pr comment` only.
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
	envFakeGhViewJSON  = "FAKE_GH_VIEW_JSON"
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
		// Default empty view object; status leaves override FAKE_GH_VIEW_JSON.
		envFakeGhViewJSON+"={}",
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
if [ "$1" = "pr" ] && [ "$2" = "view" ]; then
  # P4: gh pr view <n> --json number,title,url,state,isDraft,reviewDecision,statusCheckRollup
  printf '%s\n' "${FAKE_GH_VIEW_JSON:-{}}"
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

// setFakeGhViewJSON replaces or appends FAKE_GH_VIEW_JSON for status-mode
// `gh pr view` responses.
func setFakeGhViewJSON(t *testing.T, req *Request, js string) {
	t.Helper()
	if js == "" {
		js = "{}"
	}
	replaced := false
	for i, e := range req.ExtraEnv {
		if strings.HasPrefix(e, envFakeGhViewJSON+"=") {
			req.ExtraEnv[i] = envFakeGhViewJSON + "=" + js
			replaced = true
			break
		}
	}
	if !replaced {
		req.ExtraEnv = append(req.ExtraEnv, envFakeGhViewJSON+"="+js)
	}
}

// prViewJSON builds a compact gh pr view --json object for status leaves.
// rollupJSON is a raw JSON array/null for statusCheckRollup (not re-quoted).
// reviewDecision examples: "REVIEW_REQUIRED", "APPROVED", "CHANGES_REQUESTED", "".
func prViewJSON(number int, title, url, state string, isDraft bool, reviewDecision, rollupJSON string) string {
	if number == 0 {
		number = prExistingNumber
	}
	if title == "" {
		title = prExistingTitle
	}
	if url == "" {
		url = prDefaultURL
	}
	if state == "" {
		state = "OPEN"
	}
	if rollupJSON == "" {
		rollupJSON = "[]"
	}
	return fmt.Sprintf(
		`{"number":%d,"title":%q,"url":%q,"state":%q,"isDraft":%t,"reviewDecision":%q,"statusCheckRollup":%s}`,
		number, title, url, state, isDraft, reviewDecision, rollupJSON,
	)
}

// Rollup fixture fragments for statusCheckRollup (product maps → success|failure|pending|none|mixed).
const (
	// Single completed SUCCESS check → Checks: success
	prRollupSuccessJSON = `[{"__typename":"CheckRun","name":"ci","status":"COMPLETED","conclusion":"SUCCESS","state":"SUCCESS"}]`
	// Single completed FAILURE check → Checks: failure
	prRollupFailureJSON = `[{"__typename":"CheckRun","name":"ci","status":"COMPLETED","conclusion":"FAILURE","state":"FAILURE"}]`
	// In-progress check, no failure → Checks: pending
	prRollupPendingJSON = `[{"__typename":"CheckRun","name":"ci","status":"IN_PROGRESS","conclusion":"","state":"PENDING"}]`
	// Empty rollup → Checks: none
	prRollupNoneJSON = `[]`
)

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

// prArgs builds the standard successful create/attach --pr argv.
func prArgs(title, comment string) []string {
	return []string{"--pr", "--title", title, "--comment", comment}
}

func prDefaultArgs() []string {
	return prArgs(prDefaultTitle, prDefaultComment)
}

// prShowArgs is bare show mode (P1).
func prShowArgs() []string {
	return []string{"--pr"}
}

// prCommentOnlyArgs is comment-only mode (P2): --pr --comment C, no --title.
func prCommentOnlyArgs(comment string) []string {
	if comment == "" {
		comment = prDefaultComment
	}
	return []string{"--pr", "--comment", comment}
}

// prStatusArgs is PR status mode (P4): --pr --status, no title/comment/push.
func prStatusArgs() []string {
	return []string{"--pr", "--status"}
}

// prStatusThenPrArgs: argv order free — --status before --pr.
func prStatusThenPrArgs() []string {
	return []string{"--status", "--pr"}
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

// prShowStdout is show-mode success: URL only + trailing newline.
func prShowStdout(url string) string {
	if url == "" {
		url = prDefaultURL
	}
	return url + "\n"
}

// prStatusStdout is status-mode success block (exact field labels + spacing).
// Values: state open|draft|…; checks success|failure|pending|none|mixed;
// reviews approved|changes requested|review required|none.
// Field labels left-aligned; values pad to column 11 (spaces after colon).
// Flexible whitespace is OK for implementers if labels/values match; tests pin this shape.
func prStatusStdout(url, state, title, checks, reviews string) string {
	if url == "" {
		url = prDefaultURL
	}
	if title == "" {
		title = prExistingTitle
	}
	return fmt.Sprintf("%s\nState:     %s\nTitle:     %s\nChecks:    %s\nReviews:   %s\n",
		url, state, title, checks, reviews)
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
	_ = setFakeGhViewJSON
	_ = prViewJSON
	_ = prRollupSuccessJSON
	_ = prRollupFailureJSON
	_ = prRollupPendingJSON
	_ = prRollupNoneJSON
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
	_ = prShowArgs
	_ = prCommentOnlyArgs
	_ = prStatusArgs
	_ = prStatusThenPrArgs
	_ = prPushConfirmLine
	_ = prCreatedStdout
	_ = prCreatedWithPushStdout
	_ = prExistingStdout
	_ = prShowStdout
	_ = prStatusStdout
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
