# Scenario

**Feature**: linked consumer `wrk --status` / `--repos` reports live `./external/*` dep worktrees

```
# layout (user-visible failure case)
consumer main  + wrk --new -> linked consumer under WRK_HOME/worktrees/…
dep main(s)    + git worktree add -> {consumerWt}/external/<name>
  -> cwd = consumerWt
  -> wrk --status  => ≥1 Dir block for consumer + one Dir block per external dep
  -> wrk --repos   => "." and "external/…" lines (same discovery)

# incomplete warm index must not hide live under-root checkouts
FakeHome + seed home/repos.json listing only consumerWt
  -> warm serve + young external/* still must surface external blocks
```

## Preconditions

- Git available (parent status Setup).
- Isolated `WRK_HOME` at `{WorkRoot}/.wrk`; `FakeHome = WorkRoot` so product
  scan cache lands under `{FakeHome}/.cache/git-repo-scan` (Capture sets `HOME`
  under mutex — Parallel-safe, no harness `t.Setenv`/`t.Chdir`).
- External dep checkouts are **linked worktrees of other mains** (not consumer-owned),
  created via `git -C <depMain> worktree add` into `{consumerWt}/external/…`.
- L2: leaves set `req.InProcess = true` (`wrkcli.Capture` via mega-tree `Run`).

## Steps

1. Branch Setup registers helpers and default `Args = ["--status"]`.
2. Descendant leaves build consumer linked wt ± external dep worktree(s),
   optionally seed incomplete warm cache, set `RepoDir` / `Args`.

## Context

- Linked cwd (no `--main`) uses **scan-only** status (no main primary / `---- external ----`
  section header). Blocks still include `Master:` for linked paths.
- Desired discovery: **complete under the status checkout root** (live expand /
  ListWorktrees-style completeness as implementer chooses) — must not rely on a
  stale warm index path list alone.
- Hostile seed models incomplete index: only `consumerWt` in `home/repos.json`,
  plus `meta.json` `last_scan_end=now` so sibling/walk-consume budgets are 0 and
  brand-new `external/` units stay young under default YoungAge.

## Tree Overview

```
linked-external-deps/
├── no-external/
│   └── status/                 # linked cwd, no external → single Dir: . block
├── one-external/
│   ├── status/                 # consumer + one external/* → ≥2 blocks
│   └── repos/                  # --repos: . + external/…
└── two-external/
    └── status/                 # consumer + two external/* → 3 blocks
```

Split factor (MECE, significance-first):

1. **External dep presence** under linked consumer (`none` / `one` / `two`).
2. Within `one-external`: **CLI surface** (`--status` vs `--repos` same discovery).

## Test Case Index

| # | Leaf | Description |
|---|------|-------------|
| 1 | no-external/status | Linked consumer, no `./external` → one `Dir: .` block; no section header |
| 2 | one-external/status | Linked + one dep WT under `external/` + incomplete warm seed → 2 blocks (`.` + `external/…`); Master on both; no `---- external ----` |
| 3 | one-external/repos | Same fixture; `wrk --repos` prints `.` and `external/…` |
| 4 | two-external/status | Linked + two dep WTs + incomplete warm seed → 3 blocks; both external paths present |

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/xhd2015/doctest/session"
	"github.com/xhd2015/gitops/git"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	if len(req.Args) == 0 {
		req.Args = []string{"--status"}
	}
	// Isolate product scan cache under WorkRoot (Capture HOME under mutex).
	req.FakeHome = req.WorkRoot
	ensureLinkedExternalDepsHelpersUsed()
	return nil
}

func linkedExternalInitMain(t *testing.T, path, subject string) {
	t.Helper()
	statusInitRepoWithSubject(t, path, subject)
}

// setupLinkedConsumer creates consumer main + wrk --new linked worktree.
// Sets req.MainRepo, req.WtDir, req.WtBranch, req.ConsumerTop.
func setupLinkedConsumer(t *testing.T, req *Request) (consumerMain, consumerWt, branch string) {
	t.Helper()
	consumerMain = filepath.Join(req.WorkRoot, "consumer")
	linkedExternalInitMain(t, consumerMain, "linked consumer main")
	// go.mod so wrk --new follows go-module create path consistently with bring fixtures.
	writeFile(t, filepath.Join(consumerMain, "go.mod"), "module example.com/consumer\n\ngo 1.22\n")
	runGitIsolated(t, consumerMain, "add", "go.mod")
	runGitIsolated(t, consumerMain, "commit", "-m", "add go.mod")

	consumerWt = runWrkFrom(t, req, consumerMain)
	branch = branchName("main", wrkDate, 0)
	req.MainRepo = consumerMain
	req.WtDir = consumerWt
	req.WtBranch = branch
	req.ConsumerTop = consumerWt
	return consumerMain, consumerWt, branch
}

// setupDepMain creates an independent dep main repo (not under consumer).
func setupDepMain(t *testing.T, workRoot, name, subject string) string {
	t.Helper()
	dep := filepath.Join(workRoot, name)
	linkedExternalInitMain(t, dep, subject)
	if resolved, err := filepath.EvalSymlinks(dep); err == nil {
		dep = resolved
	}
	return dep
}

// addExternalDepWorktree registers a linked worktree of depMain under
// {consumerWt}/external/{relName} on a new branch.
func addExternalDepWorktree(t *testing.T, consumerWt, depMain, relName, branch string) string {
	t.Helper()
	extDir := filepath.Join(consumerWt, "external", filepath.FromSlash(relName))
	mkdirAll(t, filepath.Dir(extDir))
	runGitIsolated(t, depMain, "worktree", "add", "-b", branch, extDir)
	return extDir
}

// ignoreExternalOnConsumer commits /external (and optional extra patterns) so
// the consumer linked wt porcelain stays clean while nested deps live on disk.
func ignoreExternalOnConsumer(t *testing.T, consumerWt string, extraPatterns ...string) {
	t.Helper()
	patterns := append([]string{"/external"}, extraPatterns...)
	writeFile(t, filepath.Join(consumerWt, ".gitignore"), strings.Join(patterns, "\n")+"\n")
	runGitIsolated(t, consumerWt, "add", ".gitignore")
	runGitIsolated(t, consumerWt, "commit", "-m", "ignore external dep worktrees")
}

// seedIncompleteConsumerOnlyIndex writes a warm-eligible home universe index that
// lists only consumerWt (omitting live ./external/*). Also stamps meta.json
// last_scan_end=now so sibling probe + walk-log consume budgets are 0, and
// brand-new external/ units stay young under default YoungAge.
// Pins: status/repos discovery must not trust incomplete warm path rows alone.
func seedIncompleteConsumerOnlyIndex(t *testing.T, fakeHome, consumerWt string) {
	t.Helper()
	if fakeHome == "" {
		t.Fatal("seedIncompleteConsumerOnlyIndex requires FakeHome")
	}
	consumerWt = statusNormalizePath(t, consumerWt)
	gitDir := linkedExternalGitDir(t, consumerWt)
	now := time.Now().UTC().Format(time.RFC3339)

	cacheRoot := filepath.Join(fakeHome, ".cache", "git-repo-scan")
	homeDir := filepath.Join(cacheRoot, "home")
	mkdirAll(t, homeDir)

	type indexEntry struct {
		Path     string `json:"path"`
		RepoType string `json:"repo_type"`
		GitDir   string `json:"git_dir"`
		Depth    int    `json:"depth"`
		SeenAt   string `json:"seen_at"`
	}
	type indexDoc struct {
		Version   int          `json:"version"`
		Universe  string       `json:"universe"`
		Base      string       `json:"base"`
		UpdatedAt string       `json:"updated_at"`
		Repos     []indexEntry `json:"repos"`
	}
	doc := indexDoc{
		Version:   1,
		Universe:  "home",
		Base:      consumerWt,
		UpdatedAt: now,
		Repos: []indexEntry{{
			Path:     consumerWt,
			RepoType: "worktree",
			GitDir:   gitDir,
			Depth:    0,
			SeenAt:   now,
		}},
	}
	data, err := json.Marshal(doc)
	if err != nil {
		t.Fatalf("marshal incomplete repos.json: %v", err)
	}
	writeFile(t, filepath.Join(homeDir, "repos.json"), string(data))

	meta := map[string]string{"last_scan_end": now}
	metaData, err := json.Marshal(meta)
	if err != nil {
		t.Fatalf("marshal meta.json: %v", err)
	}
	writeFile(t, filepath.Join(homeDir, "meta.json"), string(metaData))
}

func linkedExternalGitDir(t *testing.T, checkout string) string {
	t.Helper()
	gitPath := filepath.Join(checkout, ".git")
	info, err := os.Stat(gitPath)
	if err != nil {
		t.Fatalf("stat %s: %v", gitPath, err)
	}
	if info.IsDir() {
		return statusNormalizePath(t, gitPath)
	}
	raw, err := os.ReadFile(gitPath)
	if err != nil {
		t.Fatalf("read %s: %v", gitPath, err)
	}
	s := strings.TrimSpace(string(raw))
	const prefix = "gitdir: "
	if !strings.HasPrefix(s, prefix) {
		t.Fatalf("unexpected .git file in %s: %q", checkout, s)
	}
	return statusNormalizePath(t, strings.TrimSpace(s[len(prefix):]))
}

func masterBriefFromResult(result *git.CompareBranchesResult) string {
	switch result.Relation {
	case git.BranchRelationSame:
		return "identical"
	case git.BranchRelationAIsAncestorOfB:
		commitWord := "commit"
		if result.CommitsAheadB != 1 {
			commitWord = "commits"
		}
		return fmt.Sprintf("needs merge back(+%d %s)", result.CommitsAheadB, commitWord)
	case git.BranchRelationBIsAncestorOfA:
		commitWord := "commit"
		if result.CommitsAheadA != 1 {
			commitWord = "commits"
		}
		return fmt.Sprintf("needs fast forward(+%d %s)", result.CommitsAheadA, commitWord)
	case git.BranchRelationDiverged:
		diverged := result.CommitsAheadA + result.CommitsAheadB
		commitWord := "commit"
		if diverged != 1 {
			commitWord = "commits"
		}
		return fmt.Sprintf("diverged(%d %s)", diverged, commitWord)
	default:
		return fmt.Sprintf("unknown branch relation %v", result.Relation)
	}
}

func masterField(t *testing.T, mainRepo, mainBranch, wtBranch string) string {
	t.Helper()
	result, err := git.CompareBranches(mainRepo, mainBranch, wtBranch)
	if err != nil {
		t.Fatalf("CompareBranches(%q, %q, %q): %v", mainRepo, mainBranch, wtBranch, err)
	}
	return "Master:       " + masterBriefFromResult(result)
}

// linkedScanBlock builds one scan-phase linked block with Master (no Remote).
func linkedScanBlock(t *testing.T, invCwd, mainRepo, repoDir, wtBranch, statusLine string) string {
	t.Helper()
	return fmt.Sprintf("Dir:          %s\n%s\n%s\nStatus:       %s\n%s",
		statusDirLine(t, invCwd, repoDir),
		statusBranchLine(t, repoDir),
		statusCommitLine(t, repoDir),
		statusLine,
		masterField(t, mainRepo, "main", wtBranch),
	)
}

func assertStdoutHasDirLine(t *testing.T, stdout, dirLine string) {
	t.Helper()
	want := "Dir:          " + dirLine
	if !strings.Contains(stdout, want) {
		t.Fatalf("stdout missing %q, got:\n%s", want, stdout)
	}
}

func ensureLinkedExternalDepsHelpersUsed() {
	_ = setupLinkedConsumer
	_ = setupDepMain
	_ = addExternalDepWorktree
	_ = ignoreExternalOnConsumer
	_ = seedIncompleteConsumerOnlyIndex
	_ = linkedExternalGitDir
	_ = masterField
	_ = linkedScanBlock
	_ = assertStdoutHasDirLine
}
```
