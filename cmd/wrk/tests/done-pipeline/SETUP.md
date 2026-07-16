# Scenario

**Feature**: after successful `--done`, optional post steps run in fixed order: sync → tag-next → push → propagate-tags

```
# primary --done succeeds (not aborted) then ordered post-pipeline
linked wt (ahead) [+ optional wtB] [+ bare origin] [+ registered consumer]
  -> wrk --done -y [--sync] [--tag-next] [--push] [--propagate-tags]
  -> merge-back --rm (message on stdout)
  -> blank line + runSync(main)? when --sync
  -> blank line + tag-next apply on main (local tags)? when --tag-next
  -> blank line + runPushMain(main, tags=created)? when --push
  -> blank line + runPropagateTags(main, WRK_HOME)? when --propagate-tags
       (uses new tags after tag-next, or existing source tags without --tag-next;
        dry-run threads planned next tags into propagate plan)
  -> event command stays "done"
```

## Preconditions

- Git available; monotree root helpers (`setupWrkWorktreeFromMain`, `setupCompositionTwoWTs`,
  `commitAheadOnWorktree`, `primaryThenSyncStdout`, `v2StdoutTemplate`, …).
- **Real apply** post-pipeline after done (composition dry-run lives under `dry-run/`; merge-back twin under `merge-back-pipeline/`).
- **P7 propagate leaves** need Go toolchain + offline `file://` module proxy (ExtraEnv) so
  consumer `go mod tidy` / `go build` succeed without network.
- Locked behavior (docs + GREEN leaves for sync/tag/push; Classic RED for propagate until P7):
  1. dispatch prefers **done** over bare `runTagNext` / `runPropagateTags` when primary set,
  2. `runDone` post order: sync → tag-next → push → **propagate-tags**,
  3. with `--push`, `runPushMain(..., createdTags)` pushes branch + tags,
  4. `--propagate-tags` may compose with primary (± `--tag-next`); event stays `"done"`.

## Steps

- Grouping only: leaves call fixture helpers and set `req.Args` / `req.StdinInput`.

```go
import (
	"archive/zip"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

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

// --- P7: primary + --propagate-tags multi-project fixtures ---

// Contract versions for propagate leaves (match root-bump tag-next v0.0.1 → v0.0.2).
const (
	pipelinePropOldTag     = "v0.0.1"
	pipelinePropNextTag    = "v0.0.2"
	pipelinePropModulePath = "example.com/lib"
	pipelinePropAppModule  = "example.com/app"
)

type pipelineProjectsJSONEntry struct {
	Path    string `json:"path"`
	AddedAt string `json:"added_at"`
	Source  string `json:"source"`
}

type pipelineProjectsJSONFile struct {
	Version  int                         `json:"version"`
	Projects []pipelineProjectsJSONEntry `json:"projects"`
}

func pipelineWriteProjectsJSON(t *testing.T, wrkHome string, paths ...string) {
	t.Helper()
	var projects []pipelineProjectsJSONEntry
	for _, p := range paths {
		projects = append(projects, pipelineProjectsJSONEntry{
			Path:    p,
			AddedAt: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			Source:  "manual",
		})
	}
	pf := pipelineProjectsJSONFile{Version: 1, Projects: projects}
	data, err := json.MarshalIndent(pf, "", "  ")
	if err != nil {
		t.Fatalf("marshal projects.json: %v", err)
	}
	if err := os.MkdirAll(wrkHome, 0o755); err != nil {
		t.Fatalf("mkdir WRK_HOME: %v", err)
	}
	path := filepath.Join(wrkHome, "projects.json")
	if err := os.WriteFile(path, append(data, '\n'), 0o644); err != nil {
		t.Fatalf("write projects.json: %v", err)
	}
}

func pipelineInitGitRepo(t *testing.T, path string) {
	t.Helper()
	mkdirAll(t, path)
	runGitIsolated(t, path, "-c", "init.templateDir=", "init", "-b", "main")
	runGitIsolated(t, path, "config", "user.email", "test@test.com")
	runGitIsolated(t, path, "config", "user.name", "Test")
}

func pipelineWriteGoMod(t *testing.T, dir, modulePath string, requires []string) {
	t.Helper()
	var b strings.Builder
	b.WriteString("module ")
	b.WriteString(modulePath)
	b.WriteString("\n\ngo 1.22\n")
	if len(requires) > 0 {
		b.WriteString("\nrequire (\n")
		for _, r := range requires {
			parts := strings.SplitN(r, "@", 2)
			if len(parts) != 2 {
				t.Fatalf("require %q must be path@version", r)
			}
			fmt.Fprintf(&b, "\t%s %s\n", parts[0], parts[1])
		}
		b.WriteString(")\n")
	}
	writeFile(t, filepath.Join(dir, "go.mod"), b.String())
}

func pipelineWriteConsumerMain(t *testing.T, dir string, importPaths ...string) {
	t.Helper()
	var b strings.Builder
	b.WriteString("package main\n\n")
	if len(importPaths) > 0 {
		b.WriteString("import (\n")
		for _, p := range importPaths {
			fmt.Fprintf(&b, "\t_ %q\n", p)
		}
		b.WriteString(")\n\n")
	}
	b.WriteString("func main() {}\n")
	writeFile(t, filepath.Join(dir, "main.go"), b.String())
}

func pipelineLibGo(version string) string {
	return "package lib\n\nfunc Version() string { return \"" + version + "\" }\n"
}

func pipelineReadFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func pipelineGoModPath(repo string) string {
	return filepath.Join(repo, "go.mod")
}

func pipelineSeedFileModuleProxy(t *testing.T, proxyRoot, modulePath, version, srcDir string) {
	t.Helper()
	vDir := filepath.Join(append([]string{proxyRoot}, strings.Split(modulePath, "/")...)...)
	vDir = filepath.Join(vDir, "@v")
	mkdirAll(t, vDir)

	modContent := pipelineReadFile(t, filepath.Join(srcDir, "go.mod"))
	writeFile(t, filepath.Join(vDir, version+".mod"), modContent)
	writeFile(t, filepath.Join(vDir, version+".info"),
		fmt.Sprintf(`{"Version":%q,"Time":"2026-07-01T00:00:00Z"}`+"\n", version))
	listPath := filepath.Join(vDir, "list")
	existing := ""
	if data, err := os.ReadFile(listPath); err == nil {
		existing = string(data)
	}
	if !strings.Contains(existing, version) {
		writeFile(t, listPath, existing+version+"\n")
	}
	zipPath := filepath.Join(vDir, version+".zip")
	if err := pipelineWriteModuleZip(zipPath, modulePath, version, srcDir); err != nil {
		t.Fatalf("write module zip %s: %v", zipPath, err)
	}
}

func pipelineWriteModuleZip(zipPath, modulePath, version, srcDir string) error {
	f, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer f.Close()
	zw := zip.NewWriter(f)
	prefix := modulePath + "@" + version + "/"
	err = filepath.Walk(srcDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			base := filepath.Base(path)
			if path != srcDir && base == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		rel, err := filepath.Rel(srcDir, path)
		if err != nil {
			return err
		}
		if strings.HasPrefix(rel, ".git"+string(filepath.Separator)) || rel == ".git" {
			return nil
		}
		rel = filepath.ToSlash(rel)
		w, err := zw.Create(prefix + rel)
		if err != nil {
			return err
		}
		in, err := os.Open(path)
		if err != nil {
			return err
		}
		_, copyErr := io.Copy(w, in)
		_ = in.Close()
		return copyErr
	})
	if err != nil {
		_ = zw.Close()
		return err
	}
	return zw.Close()
}

func pipelineEnableFileModuleProxy(t *testing.T, req *Request, proxyRoot string) {
	t.Helper()
	abs, err := filepath.Abs(proxyRoot)
	if err != nil {
		t.Fatalf("abs proxy: %v", err)
	}
	proxyURL := "file://" + abs
	req.ExtraEnv = append(req.ExtraEnv,
		"GOPROXY="+proxyURL,
		"GOSUMDB=off",
		"GONOSUMDB=*",
	)
}

func pipelineSeedOldModuleProxyFromContent(t *testing.T, proxyRoot, modulePath, version, libGo string) {
	t.Helper()
	tmp := filepath.Join(filepath.Dir(proxyRoot), "seed-old-"+version)
	mkdirAll(t, tmp)
	pipelineWriteGoMod(t, tmp, modulePath, nil)
	writeFile(t, filepath.Join(tmp, "lib.go"), libGo)
	pipelineSeedFileModuleProxy(t, proxyRoot, modulePath, version, tmp)
}

func pipelineRequireGo(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not available")
	}
}

func pipelineSnapshotAppBaseline(t *testing.T, req *Request) {
	t.Helper()
	dir := filepath.Join(req.WorkRoot, "_prop_baseline")
	mkdirAll(t, dir)
	writeFile(t, filepath.Join(dir, "app.gomod"), pipelineReadFile(t, pipelineGoModPath(req.SecondRepo)))
	writeFile(t, filepath.Join(dir, "app.head"), revParseHEAD(t, req.SecondRepo)+"\n")
	writeFile(t, filepath.Join(dir, "main.head"), revParseHEAD(t, req.MainRepo)+"\n")
}

func pipelineReadAppBaseline(t *testing.T, req *Request, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(req.WorkRoot, "_prop_baseline", name))
	if err != nil {
		t.Fatalf("read prop baseline %s: %v", name, err)
	}
	return strings.TrimSpace(string(data))
}

// setupDonePipelinePropagateTagNext builds source lib + consumer app + linked wt ahead
// so post-merge --tag-next plans/creates v0.0.2 and --propagate-tags can bump the app.
// Fields: MainRepo=lib, SecondRepo=app, DepModulePath=example.com/lib, RepoDir=wt.
func setupDonePipelinePropagateTagNext(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)
	pipelineRequireGo(t)

	libPath := filepath.Join(req.WorkRoot, "repos", "lib")
	appPath := filepath.Join(req.WorkRoot, "repos", "app")

	pipelineInitGitRepo(t, libPath)
	pipelineWriteGoMod(t, libPath, pipelinePropModulePath, nil)
	writeFile(t, filepath.Join(libPath, "lib.go"), pipelineLibGo(pipelinePropOldTag))
	runGitIsolated(t, libPath, "add", ".")
	runGitIsolated(t, libPath, "commit", "-m", "init lib "+pipelinePropOldTag)
	createLightweightTag(t, libPath, pipelinePropOldTag, "")
	libPath = compositionResolvePath(t, libPath)
	req.MainRepo = libPath
	req.DepModulePath = pipelinePropModulePath

	pipelineInitGitRepo(t, appPath)
	pipelineWriteGoMod(t, appPath, pipelinePropAppModule, []string{
		pipelinePropModulePath + "@" + pipelinePropOldTag,
	})
	pipelineWriteConsumerMain(t, appPath, pipelinePropModulePath)
	runGitIsolated(t, appPath, "add", ".")
	runGitIsolated(t, appPath, "commit", "-m", "init app")
	appPath = compositionResolvePath(t, appPath)
	req.SecondRepo = appPath

	pipelineWriteProjectsJSON(t, req.WrkHome, libPath, appPath)

	wtDir := runWrkFrom(t, req, libPath)
	wtDir = compositionResolvePath(t, wtDir)
	req.WtDir = wtDir
	req.WtBranch = branchName("main", wrkDate, 0)

	// Owned change on worktree → after merge, tag-next root-bumps to v0.0.2.
	writeFile(t, filepath.Join(wtDir, "lib.go"), pipelineLibGo(pipelinePropNextTag))
	runGitIsolated(t, wtDir, "add", "lib.go")
	runGitIsolated(t, wtDir, "commit", "-m", "bump lib for next tag")
	// Also leave a plain feature marker (done-pipeline style).
	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead of main")

	proxyRoot := filepath.Join(req.WorkRoot, "modproxy")
	pipelineSeedFileModuleProxy(t, proxyRoot, pipelinePropModulePath, pipelinePropNextTag, wtDir)
	pipelineSeedOldModuleProxyFromContent(t, proxyRoot, pipelinePropModulePath, pipelinePropOldTag,
		pipelineLibGo(pipelinePropOldTag))
	pipelineEnableFileModuleProxy(t, req, proxyRoot)

	req.RepoDir = wtDir
	pipelineSnapshotAppBaseline(t, req)
}

// setupDonePipelinePropagateExisting: source already has NextTag; consumer still on OldTag.
// --done -y --propagate-tags (no --tag-next) bumps consumer from existing release tags.
func setupDonePipelinePropagateExisting(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)
	pipelineRequireGo(t)

	libPath := filepath.Join(req.WorkRoot, "repos", "lib")
	appPath := filepath.Join(req.WorkRoot, "repos", "app")

	pipelineInitGitRepo(t, libPath)
	pipelineWriteGoMod(t, libPath, pipelinePropModulePath, nil)
	writeFile(t, filepath.Join(libPath, "lib.go"), pipelineLibGo(pipelinePropOldTag))
	runGitIsolated(t, libPath, "add", ".")
	runGitIsolated(t, libPath, "commit", "-m", "init lib "+pipelinePropOldTag)
	createLightweightTag(t, libPath, pipelinePropOldTag, "")
	// Advance source to NextTag content and tag it (existing release).
	writeFile(t, filepath.Join(libPath, "lib.go"), pipelineLibGo(pipelinePropNextTag))
	runGitIsolated(t, libPath, "add", "lib.go")
	runGitIsolated(t, libPath, "commit", "-m", "release "+pipelinePropNextTag)
	createLightweightTag(t, libPath, pipelinePropNextTag, "")
	libPath = compositionResolvePath(t, libPath)
	req.MainRepo = libPath
	req.DepModulePath = pipelinePropModulePath

	pipelineInitGitRepo(t, appPath)
	pipelineWriteGoMod(t, appPath, pipelinePropAppModule, []string{
		pipelinePropModulePath + "@" + pipelinePropOldTag,
	})
	pipelineWriteConsumerMain(t, appPath, pipelinePropModulePath)
	runGitIsolated(t, appPath, "add", ".")
	runGitIsolated(t, appPath, "commit", "-m", "init app")
	appPath = compositionResolvePath(t, appPath)
	req.SecondRepo = appPath

	pipelineWriteProjectsJSON(t, req.WrkHome, libPath, appPath)

	wtDir := runWrkFrom(t, req, libPath)
	wtDir = compositionResolvePath(t, wtDir)
	req.WtDir = wtDir
	req.WtBranch = branchName("main", wrkDate, 0)
	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead of main")

	proxyRoot := filepath.Join(req.WorkRoot, "modproxy")
	// NextTag tree is main's tagged tip; worktree has extra feature-work only.
	pipelineSeedFileModuleProxy(t, proxyRoot, pipelinePropModulePath, pipelinePropNextTag, libPath)
	pipelineSeedOldModuleProxyFromContent(t, proxyRoot, pipelinePropModulePath, pipelinePropOldTag,
		pipelineLibGo(pipelinePropOldTag))
	pipelineEnableFileModuleProxy(t, req, proxyRoot)

	req.RepoDir = wtDir
	pipelineSnapshotAppBaseline(t, req)
}

func propDepsBumpSubject(depModule, version string) string {
	return fmt.Sprintf("chore(deps): bump %s to %s", depModule, version)
}

func propStageApplyStdout(sourceAbs, modulePath, oldTag, nextTag, appBase, appShort, subject string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "source: %s\n", sourceAbs)
	fmt.Fprintf(&b, "  %s  @ %s  (tag %s)\n\n", modulePath, nextTag, nextTag)
	fmt.Fprintf(&b, "updated %s  (project %s)\n", pipelinePropAppModule, appBase)
	fmt.Fprintf(&b, "  %s  %s -> %s\n", modulePath, oldTag, nextTag)
	b.WriteString("  go build ./... ok\n")
	fmt.Fprintf(&b, "  committed %s  %s\n\n", appShort, subject)
	b.WriteString("updated 1 module across 1 project\n")
	return b.String()
}

func propStageDryRunStdout(sourceAbs, modulePath, oldTag, nextTag, appBase string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "source: %s\n", sourceAbs)
	fmt.Fprintf(&b, "  %s  @ %s  (tag %s)\n\n", modulePath, nextTag, nextTag)
	fmt.Fprintf(&b, "would: update %s  (project %s)\n", pipelinePropAppModule, appBase)
	fmt.Fprintf(&b, "  %s  %s -> %s\n\n", modulePath, oldTag, nextTag)
	b.WriteString("would: update 1 module across 1 project\n")
	return b.String()
}

func assertGoModRequireVersion(t *testing.T, goMod, modulePath, version string) {
	t.Helper()
	for _, line := range strings.Split(goMod, "\n") {
		trim := strings.TrimSpace(line)
		if strings.HasPrefix(trim, "require ") {
			fields := strings.Fields(trim)
			if len(fields) >= 3 && fields[1] == modulePath && fields[2] == version {
				return
			}
			continue
		}
		fields := strings.Fields(trim)
		if len(fields) >= 2 && fields[0] == modulePath && fields[1] == version {
			return
		}
	}
	t.Fatalf("go.mod missing require %s %s\n%s", modulePath, version, goMod)
}

func assertAppBumpedAndCommitted(t *testing.T, req *Request, wantVersion, wantSubject string) {
	t.Helper()
	app := req.SecondRepo
	if app == "" {
		t.Fatal("SecondRepo (app) empty")
	}
	gotMod := pipelineReadFile(t, pipelineGoModPath(app))
	assertGoModRequireVersion(t, gotMod, req.DepModulePath, wantVersion)
	before := pipelineReadAppBaseline(t, req, "app.head")
	gotHEAD := revParseHEAD(t, app)
	if gotHEAD == before {
		t.Fatalf("app HEAD did not advance after propagate (still %s)", gotHEAD)
	}
	subject := strings.TrimSpace(gitOutputIsolated(t, app, "log", "-1", "--format=%s"))
	if subject != wantSubject {
		t.Fatalf("commit subject want %q got %q", wantSubject, subject)
	}
}

func assertAppUnchangedFromBaseline(t *testing.T, req *Request) {
	t.Helper()
	app := req.SecondRepo
	gotMod := strings.TrimSpace(pipelineReadFile(t, pipelineGoModPath(app)))
	wantMod := pipelineReadAppBaseline(t, req, "app.gomod")
	if gotMod != wantMod {
		t.Fatalf("app go.mod mutated\nbefore:\n%s\nafter:\n%s", wantMod, gotMod)
	}
	gotHEAD := revParseHEAD(t, app)
	wantHEAD := pipelineReadAppBaseline(t, req, "app.head")
	if gotHEAD != wantHEAD {
		t.Fatalf("app HEAD mutated: before %s after %s", wantHEAD, gotHEAD)
	}
}

func assertNoPropagateStageStdout(t *testing.T, stdout string) {
	t.Helper()
	// Propagate human blocks / apply markers must not appear when aborted or not requested.
	assertNotContains(t, stdout, "would: update ")
	assertNotContains(t, stdout, "updated 1 module")
	assertNotContains(t, stdout, "would: update 1 module")
	if strings.Contains(stdout, "source: ") && strings.Contains(stdout, "(tag ") {
		t.Fatalf("stdout looks like propagate source block; got %q", stdout)
	}
	if strings.Contains(stdout, "chore(deps):") {
		t.Fatalf("stdout must not mention chore(deps) commit when propagate skipped; got %q", stdout)
	}
}

// keep helpers referenced for inheritance compilation
var (
	_ = setupDonePipelineLocal
	_ = setupDonePipelineWithOrigin
	_ = setupDonePipelineSync
	_ = setupDonePipelineSyncWithOrigin
	_ = setupDonePipelinePropagateTagNext
	_ = setupDonePipelinePropagateExisting
	_ = joinMajorStages
	_ = tagNextRootBumpApplyStdout
	_ = donePushConfirmLine
	_ = primaryMergeMsg
	_ = assertLastEventCommandDone
	_ = assertLocalTagAtMainHEAD
	_ = assertOriginMainEqualsLocalMain
	_ = remoteTagExists
	_ = shortHEAD
	_ = propStageApplyStdout
	_ = propStageDryRunStdout
	_ = propDepsBumpSubject
	_ = assertAppBumpedAndCommitted
	_ = assertAppUnchangedFromBaseline
	_ = assertNoPropagateStageStdout
	_ = assertGoModRequireVersion
	_ = pipelineRequireGo
	_ = fmt.Sprintf
)
```
