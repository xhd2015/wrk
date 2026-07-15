# Scenario

**Feature**: after successful `--done`, optional post steps run in fixed order: sync → tag-next → push

```
# primary --done succeeds (not aborted) then ordered post-pipeline
linked wt (ahead, root-bump seed: v0.0.1) [+ optional wtB] [+ bare origin]
  -> wrk --done -y [--sync] [--tag-next] [--push]
  -> merge-back --rm (message on stdout)
  -> blank line + runSync(main)? when --sync
  -> blank line + tag-next apply on main (local tags)? when --tag-next
  -> blank line + runPushMain(main, tags=created)? when --push
  -> event command stays "done"
```

## Preconditions

- Git available; monotree root helpers (`setupWrkWorktreeFromMain`, `setupCompositionTwoWTs`,
  `commitAheadOnWorktree`, `primaryThenSyncStdout`, `v2StdoutTemplate`, …).
- **Real apply** post-pipeline after done (composition dry-run lives under `dry-run/`; merge-back twin under `merge-back-pipeline/`).
- Locked behavior (docs + GREEN leaves):
  1. dispatch prefers **done** over bare `runTagNext` when both flags set,
  2. `runDone` runs tag-next after sync and before push,
  3. with `--push`, `runPushMain(..., createdTags)` pushes branch + tags.

## Steps

- Grouping only: leaves call fixture helpers and set `req.Args` / `req.StdinInput`.

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/xhd2015/gitops/git/git_isolated"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}

func setupDonePipelineBareOrigin(t *testing.T, workRoot, name string) string {
	t.Helper()
	bare := filepath.Join(workRoot, name+".git")
	runGitIsolated(t, workRoot, "-c", "init.templateDir=", "init", "--bare", "-b", "main", bare)
	return bare
}

func createLightweightTag(t *testing.T, repo, name, ref string) {
	t.Helper()
	if ref == "" {
		ref = "HEAD"
	}
	runGitIsolated(t, repo, "tag", name, ref)
}

func tagRefExists(t *testing.T, repo, name string) bool {
	t.Helper()
	err := git_isolated.Command(repo, "rev-parse", "--verify", "refs/tags/"+name).Run()
	return err == nil
}

func remoteTagExists(t *testing.T, bareOrigin, name string) bool {
	t.Helper()
	out := gitOutputIsolated(t, bareOrigin, "show-ref", "--tags")
	prefix := "refs/tags/" + name
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		if line == "" {
			continue
		}
		fields := strings.Fields(line)
		if len(fields) >= 2 && fields[1] == prefix {
			return true
		}
	}
	return false
}

func shortHEAD(t *testing.T, repo string) string {
	t.Helper()
	return strings.TrimSpace(gitOutputIsolated(t, repo, "rev-parse", "--short=7", "HEAD"))
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

// seedMainWithRootBumpTag: main-gomod seed + lightweight v0.0.1 at HEAD.
// Post-merge owned change (feature-work) makes tag-next plan v0.0.2 (root-bump style).
func seedMainWithRootBumpTag(t *testing.T, req *Request) string {
	t.Helper()
	mainRepo := filepath.Join(req.WorkRoot, "myrepo")
	req.MainRepo = mainRepo
	cloneRepoFromSeed(t, fixtureSeedMainGoMod, buildSeedMainGoMod, mainRepo)
	mainRepo = compositionResolvePath(t, mainRepo)
	req.MainRepo = mainRepo
	createLightweightTag(t, mainRepo, "v0.0.1", "")
	return mainRepo
}

// setupDonePipelineLocal: root-bump seed + wrk wt ahead (no origin).
// Caller sets Args. RepoDir is the linked worktree.
func setupDonePipelineLocal(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)
	mainRepo := seedMainWithRootBumpTag(t, req)

	wtDir := runWrkFrom(t, req, mainRepo)
	wtDir = compositionResolvePath(t, wtDir)
	req.WtDir = wtDir
	req.WtBranch = branchName("main", wrkDate, 0)

	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead of main")
	req.RepoDir = wtDir
}

// setupDonePipelineWithOrigin: local fixture + bare origin upstream on main.
func setupDonePipelineWithOrigin(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)
	mainRepo := seedMainWithRootBumpTag(t, req)

	bare := setupDonePipelineBareOrigin(t, req.WorkRoot, "origin")
	runGitIsolated(t, mainRepo, "remote", "add", "origin", bare)
	runGitIsolated(t, mainRepo, "push", "-u", "origin", "main")
	// Also publish baseline tag so remote has lineage (branch tip is primary assert).
	runGitIsolated(t, mainRepo, "push", "origin", "v0.0.1")
	req.OriginBare = bare

	wtDir := runWrkFrom(t, req, mainRepo)
	wtDir = compositionResolvePath(t, wtDir)
	req.WtDir = wtDir
	req.WtBranch = branchName("main", wrkDate, 0)

	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead of main")
	req.RepoDir = wtDir
}

// setupDonePipelineSync: two-worktree fixture + v0.0.1 on main (no origin).
// wtA ahead (feature-work); wtB feature-stays at pre-ahead tip.
func setupDonePipelineSync(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)
	mainRepo := seedMainWithRootBumpTag(t, req)

	wtA := runWrkFrom(t, req, mainRepo)
	wtA = compositionResolvePath(t, wtA)
	req.WtDir = wtA
	req.WtBranch = branchName("main", wrkDate, 0)

	wt2Path := filepath.Join(req.WorkRoot, "wt-stays")
	runGitIsolated(t, mainRepo, "worktree", "add", "-b", "feature-stays", wt2Path)
	wt2Path = compositionResolvePath(t, wt2Path)
	req.Wt2Dir = wt2Path
	req.Wt2Branch = "feature-stays"

	commitAheadOnWorktree(t, wtA, "feature-work", "ahead of main")
	req.RepoDir = wtA
}

// setupDonePipelineSyncWithOrigin: sync fixture + bare origin.
func setupDonePipelineSyncWithOrigin(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)
	mainRepo := seedMainWithRootBumpTag(t, req)

	bare := setupDonePipelineBareOrigin(t, req.WorkRoot, "origin")
	runGitIsolated(t, mainRepo, "remote", "add", "origin", bare)
	runGitIsolated(t, mainRepo, "push", "-u", "origin", "main")
	runGitIsolated(t, mainRepo, "push", "origin", "v0.0.1")
	req.OriginBare = bare

	wtA := runWrkFrom(t, req, mainRepo)
	wtA = compositionResolvePath(t, wtA)
	req.WtDir = wtA
	req.WtBranch = branchName("main", wrkDate, 0)

	wt2Path := filepath.Join(req.WorkRoot, "wt-stays")
	runGitIsolated(t, mainRepo, "worktree", "add", "-b", "feature-stays", wt2Path)
	wt2Path = compositionResolvePath(t, wt2Path)
	req.Wt2Dir = wt2Path
	req.Wt2Branch = "feature-stays"

	commitAheadOnWorktree(t, wtA, "feature-work", "ahead of main")
	req.RepoDir = wtA
}

// joinMajorStages joins stdout stages with a blank line between each (done-sync style).
func joinMajorStages(stages ...string) string {
	var parts []string
	for _, s := range stages {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}
		parts = append(parts, s)
	}
	if len(parts) == 0 {
		return ""
	}
	return strings.Join(parts, "\n\n") + "\n"
}

// tagNextRootBumpApplyStdout is the human apply block for root v0.0.1 → v0.0.2.
func tagNextRootBumpApplyStdout(short string) string {
	return fmt.Sprintf(
		"v0.0.1        owned changed                  ->  v0.0.2\ntagged v0.0.2 @ %s\n1 tag created\n",
		short,
	)
}

func donePushConfirmLine() string {
	return "pushed main → origin/main\n"
}

func primaryMergeMsg(wtBranch string) string {
	return fmt.Sprintf("merged branch %s into main\n", wtBranch)
}

// --- events.jsonl helpers (command stays "done" under composition) ---

type pipelineWrkEvent struct {
	TS       string   `json:"ts"`
	Command  string   `json:"command"`
	WorkDir  string   `json:"work_dir"`
	MainRepo string   `json:"main_repo"`
	Args     []string `json:"args"`
	ExitCode int      `json:"exit_code"`
}

func pipelineEventsPath(wrkHome string) string {
	return filepath.Join(wrkHome, "events.jsonl")
}

func readPipelineEvents(t *testing.T, wrkHome string) []pipelineWrkEvent {
	t.Helper()
	data, err := os.ReadFile(pipelineEventsPath(wrkHome))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read events.jsonl: %v", err)
	}
	var events []pipelineWrkEvent
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var ev pipelineWrkEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			t.Fatalf("parse event line %q: %v", line, err)
		}
		events = append(events, ev)
	}
	return events
}

func assertLastEventCommandDone(t *testing.T, wrkHome string) {
	t.Helper()
	events := readPipelineEvents(t, wrkHome)
	if len(events) == 0 {
		t.Fatal("expected at least one events.jsonl entry")
	}
	ev := events[len(events)-1]
	if ev.Command != "done" {
		t.Fatalf("event command: want %q, got %q (args=%v)", "done", ev.Command, ev.Args)
	}
}

func assertLocalTagAtMainHEAD(t *testing.T, mainRepo, tag string) {
	t.Helper()
	if !tagRefExists(t, mainRepo, tag) {
		t.Fatalf("local tag %s should exist after done --tag-next", tag)
	}
	got := revParseRef(t, mainRepo, tag)
	head := revParseHEAD(t, mainRepo)
	if got != head {
		t.Fatalf("%s should point at main HEAD: tag=%s head=%s", tag, got, head)
	}
}

// keep helpers referenced for inheritance compilation
var (
	_ = setupDonePipelineLocal
	_ = setupDonePipelineWithOrigin
	_ = setupDonePipelineSync
	_ = setupDonePipelineSyncWithOrigin
	_ = joinMajorStages
	_ = tagNextRootBumpApplyStdout
	_ = donePushConfirmLine
	_ = primaryMergeMsg
	_ = assertLastEventCommandDone
	_ = assertLocalTagAtMainHEAD
	_ = assertOriginMainEqualsLocalMain
	_ = remoteTagExists
	_ = shortHEAD
	_ = fmt.Sprintf
)
```
