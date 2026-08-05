# Scenario

**Feature**: `wrk --where --pr <full-github-pr-url>` prints local worktree path(s) for the PR head branch

```
# compose (--where Bool + --pr Bool + exactly one full GitHub PR URL positional)
wrk --where --pr https://github.com/<owner>/<repo>/pull/<N>
  -> parse full URL → owner/repo/N
  -> require gh; gh pr view N --repo owner/repo --json headRefName
  -> find mains: projects.json origin match + cwd main when origin matches
  -> worktrees on headRefName (drop dead paths)
  -> 1+ paths: exit 0, abs paths lex-sorted one per line, empty stderr
  -> 0 paths / no project: non-zero, empty stdout, stderr names PR + head + repo

# full URL only; closed/merged still resolve; no --cd --pr
# flag order free; still exclusive with --status/--main/--list/…
```

## Preconditions

- Monotree root harness (`Request` / `Response` / `Run`, git helpers).
- Parent `where/` helpers (`recordSavedProject`, `initNeutralCwd`, `resolvePath`, `sortedSavedPaths`).
- L2 in-process (`req.InProcess = true`) with local fake `gh` via `PathPrepend` (mirrors `pr/` patterns; helpers live here — sibling `pr/` package is not an ancestor).
- Hermetic origin: fetch URL `https://github.com/acme/app.git`; optional local bare for pushurl (not required for where-pr read path).
- Classic TDD: product currently treats `--where` and `--pr` as mutually exclusive — leaves must **RED** until compose is implemented.
- Parallel-safe: no `t.Setenv` / `os.Setenv` / `Chdir` for isolation; use `req.ExtraEnv`, `req.PathPrepend`, absolute paths.

## Steps

- Grouping/leaves seed git mains + linked worktrees, install fake gh, set `req.Args` / `req.RepoDir`.

## Context

- PR ref is **full URL only** (`https://github.com/owner/repo/pull/N`); bare number / shorthand / scheme-less rejected.
- `--pr` remains **Bool**; URL is the single remaining positional when composed with `--where`.
- Multi-match across clones/mains: all live head checkouts, lex-sorted.
- Event command stays **`where`** when `--where` is set.
- Paths on stdout: no ANSI; trailing `\n`.

```go
import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/session"
)

const (
	wherePrOwner        = "acme"
	wherePrRepo         = "app"
	wherePrNumber       = 42
	wherePrHeadBranch   = "feature-pr"
	wherePrGithubOrigin = "https://github.com/acme/app.git"
	wherePrURL          = "https://github.com/acme/app/pull/42"
	wherePrURLTrailing  = "https://github.com/acme/app/pull/42/"

	envWherePrGhLog      = "FAKE_GH_LOG"
	envWherePrGhViewJSON = "FAKE_GH_VIEW_JSON"
	envWherePrGhViewExit = "FAKE_GH_VIEW_EXIT"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	skipIfNoGit(t)
	return nil
}

// --- fake gh (PathPrepend; local to where/pr) ---

func installWherePrFakeGh(t *testing.T, req *Request) {
	t.Helper()
	binDir := filepath.Join(req.WorkRoot, "where-pr-gh-bin")
	mkdirAll(t, binDir)
	req.PathPrepend = binDir

	logPath := filepath.Join(req.WorkRoot, "where-pr-fake-gh.log")
	req.InterceptorLog = logPath
	writeFile(t, logPath, "")

	// Default view: OPEN PR with headRefName for location lookup.
	defaultView := wherePrViewJSON(wherePrHeadBranch, "OPEN")
	req.ExtraEnv = append(req.ExtraEnv,
		envWherePrGhLog+"="+logPath,
		envWherePrGhViewJSON+"="+defaultView,
		envWherePrGhViewExit+"=0",
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

if [ "$1" = "pr" ] && [ "$2" = "view" ]; then
  code="${FAKE_GH_VIEW_EXIT:-0}"
  if [ "$code" != "0" ]; then
    printf 'gh: could not view pull request\n' >&2
    exit "$code"
  fi
  printf '%s\n' "${FAKE_GH_VIEW_JSON:-{}}"
  exit 0
fi
if [ "$1" = "pr" ] && [ "$2" = "list" ]; then
  printf '[]\n'
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

// installWherePrPathWithoutGh forces PATH to git-only (no gh). Prefer InProcess=false
// so stripped PATH is child-only Env (same rationale as pr/refuse/missing-gh).
func installWherePrPathWithoutGh(t *testing.T, req *Request) {
	t.Helper()
	binDir := filepath.Join(req.WorkRoot, "where-pr-no-gh-bin")
	mkdirAll(t, binDir)

	gitPath, err := exec.LookPath("git")
	if err != nil {
		t.Fatalf("LookPath git: %v", err)
	}
	link := filepath.Join(binDir, "git")
	if err := os.Symlink(gitPath, link); err != nil {
		t.Fatalf("symlink git: %v", err)
	}

	req.PathPrepend = ""
	req.ExtraEnv = append(req.ExtraEnv, "PATH="+binDir)
}

func setWherePrViewJSON(t *testing.T, req *Request, js string) {
	t.Helper()
	if js == "" {
		js = "{}"
	}
	replaced := false
	for i, e := range req.ExtraEnv {
		if strings.HasPrefix(e, envWherePrGhViewJSON+"=") {
			req.ExtraEnv[i] = envWherePrGhViewJSON + "=" + js
			replaced = true
			break
		}
	}
	if !replaced {
		req.ExtraEnv = append(req.ExtraEnv, envWherePrGhViewJSON+"="+js)
	}
}

func setWherePrViewExit(t *testing.T, req *Request, code int) {
	t.Helper()
	val := fmt.Sprintf("%s=%d", envWherePrGhViewExit, code)
	replaced := false
	for i, e := range req.ExtraEnv {
		if strings.HasPrefix(e, envWherePrGhViewExit+"=") {
			req.ExtraEnv[i] = val
			replaced = true
			break
		}
	}
	if !replaced {
		req.ExtraEnv = append(req.ExtraEnv, val)
	}
}

// wherePrViewJSON builds gh pr view --json headRefName[,state] for where-pr.
func wherePrViewJSON(headRefName, state string) string {
	if headRefName == "" {
		headRefName = wherePrHeadBranch
	}
	if state == "" {
		state = "OPEN"
	}
	return fmt.Sprintf(`{"headRefName":%q,"state":%q,"number":%d}`, headRefName, state, wherePrNumber)
}

// --- git fixtures ---

func wherePrSetupBareOrigin(t *testing.T, workRoot, name string) string {
	t.Helper()
	bare := filepath.Join(workRoot, name+".git")
	runGitIsolated(t, workRoot, "-c", "init.templateDir=", "init", "--bare", "-b", "main", bare)
	return bare
}

func wherePrConfigureGithubOrigin(t *testing.T, repo, bare string) {
	t.Helper()
	runGitIsolated(t, repo, "remote", "add", "origin", wherePrGithubOrigin)
	if bare != "" {
		runGitIsolated(t, repo, "config", "remote.origin.pushurl", bare)
	}
}

// wherePrSetupMainWithOrigin: main checkout + github-shaped origin under workRoot/name.
func wherePrSetupMainWithOrigin(t *testing.T, req *Request, name string) string {
	t.Helper()
	skipIfNoGit(t)
	mainRepo := filepath.Join(req.WorkRoot, name)
	initGitRepoOnMain(t, mainRepo)
	writeFile(t, filepath.Join(mainRepo, "README.md"), "# "+name+"\n")
	runGitIsolated(t, mainRepo, "add", "README.md")
	runGitIsolated(t, mainRepo, "commit", "-m", "init")
	mainRepo = compositionResolvePath(t, mainRepo)

	bare := wherePrSetupBareOrigin(t, req.WorkRoot, name+"-origin")
	wherePrConfigureGithubOrigin(t, mainRepo, bare)
	runGitIsolated(t, mainRepo, "push", "-u", "origin", "main")
	return mainRepo
}

// wherePrAddLinkedOnHead creates a linked worktree on wherePrHeadBranch under
// workRoot/linkName and returns its resolved path.
func wherePrAddLinkedOnHead(t *testing.T, req *Request, mainRepo, linkName string) string {
	t.Helper()
	linked := filepath.Join(req.WorkRoot, linkName)
	runGitIsolated(t, mainRepo, "worktree", "add", "-b", wherePrHeadBranch, linked)
	linked = compositionResolvePath(t, linked)
	writeFile(t, filepath.Join(linked, "feature.txt"), "from linked wt\n")
	runGitIsolated(t, linked, "add", "feature.txt")
	runGitIsolated(t, linked, "commit", "-m", "feature work")
	return linked
}

// wherePrSetupRecordedLinked: main in projects.json + linked wt on head; neutral cwd.
// Sets MainRepo, WtDir, RepoDir (neutral), installs fake gh with default view JSON.
func wherePrSetupRecordedLinked(t *testing.T, req *Request) {
	t.Helper()
	mainRepo := wherePrSetupMainWithOrigin(t, req, "myrepo")
	linked := wherePrAddLinkedOnHead(t, req, mainRepo, "linked-wt")
	recordSavedProject(t, req, mainRepo)

	req.MainRepo = mainRepo
	req.WtDir = linked
	req.WtBranch = wherePrHeadBranch
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")

	installWherePrFakeGh(t, req)
}

// wherePrSetupUnrecordedLinked: same git layout but does NOT pre-record projects.json.
// Cwd is the linked worktree so product may use cwd main even if not yet recorded.
func wherePrSetupUnrecordedLinked(t *testing.T, req *Request) {
	t.Helper()
	mainRepo := wherePrSetupMainWithOrigin(t, req, "myrepo")
	linked := wherePrAddLinkedOnHead(t, req, mainRepo, "linked-wt")

	req.MainRepo = mainRepo
	req.WtDir = linked
	req.WtBranch = wherePrHeadBranch
	req.RepoDir = linked

	installWherePrFakeGh(t, req)
}

// wherePrSetupTwoClonesOnHead: two main clones of acme/app, each with a linked wt
// on the same head branch name. Both recorded. Neutral cwd. Paths on WtDir + Wt2Dir.
func wherePrSetupTwoClonesOnHead(t *testing.T, req *Request) {
	t.Helper()
	mainA := wherePrSetupMainWithOrigin(t, req, "clone-aaa")
	// Second clone needs its own branch create — same branch name is OK across clones.
	mainZ := wherePrSetupMainWithOrigin(t, req, "clone-zzz")

	linkedA := wherePrAddLinkedOnHead(t, req, mainA, "wt-aaa")
	linkedZ := wherePrAddLinkedOnHead(t, req, mainZ, "wt-zzz")

	recordSavedProject(t, req, mainA)
	recordSavedProject(t, req, mainZ)

	req.MainRepo = mainA
	req.SecondRepo = mainZ
	req.WtDir = linkedA
	req.Wt2Dir = linkedZ
	req.WtBranch = wherePrHeadBranch
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")

	installWherePrFakeGh(t, req)
}

// wherePrSetupRecordedMainOnly: matching origin main recorded; no worktree on head.
func wherePrSetupRecordedMainOnly(t *testing.T, req *Request) {
	t.Helper()
	mainRepo := wherePrSetupMainWithOrigin(t, req, "myrepo")
	recordSavedProject(t, req, mainRepo)

	req.MainRepo = mainRepo
	req.WtBranch = wherePrHeadBranch
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")

	installWherePrFakeGh(t, req)
}

// wherePrSetupNoMatchingProject: neutral cwd; optional unrelated non-github project.
func wherePrSetupNoMatchingProject(t *testing.T, req *Request) {
	t.Helper()
	// Unrelated local git project without github origin matching acme/app.
	other := filepath.Join(req.WorkRoot, "other-proj")
	initGitRepoOnMain(t, other)
	writeFile(t, filepath.Join(other, "README.md"), "# other\n")
	runGitIsolated(t, other, "add", "README.md")
	runGitIsolated(t, other, "commit", "-m", "init")
	other = compositionResolvePath(t, other)
	// non-github origin
	bare := wherePrSetupBareOrigin(t, req.WorkRoot, "other-origin")
	runGitIsolated(t, other, "remote", "add", "origin", bare)
	recordSavedProject(t, req, other)

	req.MainRepo = other
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	installWherePrFakeGh(t, req)
}

// --- argv builders ---

func wherePrArgs(url string) []string {
	if url == "" {
		url = wherePrURL
	}
	return []string{"--where", "--pr", url}
}

func wherePrThenWhereArgs(url string) []string {
	if url == "" {
		url = wherePrURL
	}
	return []string{"--pr", "--where", url}
}

// wherePrURLFirstArgs: URL positional then flags (wrk URL --where --pr).
// Harness: TargetDir is first positional; Args are remaining flags.
func setWherePrURLFirst(req *Request, url string) {
	if url == "" {
		url = wherePrURL
	}
	req.TargetDir = url
	req.Args = []string{"--where", "--pr"}
}

```
