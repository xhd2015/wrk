# Scenario

**Feature**: after successful `--merge-back`, optional post steps run in fixed order: sync → tag-next → push → propagate-tags (source worktree kept)

```
# primary --merge-back succeeds (not aborted) then ordered post-pipeline; Remove=false
linked wt (ahead) [+ optional wtB] [+ bare origin] [+ registered consumer]
  -> wrk --merge-back -y [--sync] [--tag-next] [--push] [--propagate-tags]
  -> merge without remove (message on stdout; no "worktree removed:")
  -> blank line + runSync(main)? when --sync
  -> blank line + tag-next apply on main (local tags)? when --tag-next
  -> blank line + runPushMain(main, tags=created)? when --push
  -> blank line + runPropagateTags(main, WRK_HOME)? when --propagate-tags
  -> source wt stays on disk; branch kept
  -> event command stays "merge-back"
```

## Preconditions

- Git available; monotree root helpers (`setupWrkWorktreeFromMain`, `setupCompositionTwoWTs`,
  `commitAheadOnWorktree`, `primaryThenSyncStdout`, `v2StdoutTemplate`, …).
- **Real apply** post-pipeline after merge-back (composition dry-run under `dry-run/`; done twin under `done-pipeline/`).
- Reuses done-pipeline fixture shape (root-bump seed, bare origin, two-wt sync) but
  **Remove false → worktree remains**.
- **P7 propagate leaf** needs Go + offline module proxy (ExtraEnv).
- Locked behavior (docs + GREEN leaves for sync/tag/push; Classic RED for propagate until P7):
  1. dispatch prefers **merge-back** over bare `runTagNext` / `runPropagateTags` when primary set,
  2. `runMergeBack` post order: sync → tag-next → push → **propagate-tags**,
  3. with `--push`, `runPushMain(..., createdTags)` pushes branch + tags,
  4. source worktree is never removed on this path,
  5. event command stays `"merge-back"`.

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
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}

func setupMergeBackPipelineBareOrigin(t *testing.T, workRoot, name string) string {
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
	cloneMainGoModFromSeed(t, mainRepo)
	mainRepo = compositionResolvePath(t, mainRepo)
	req.MainRepo = mainRepo
	createLightweightTag(t, mainRepo, "v0.0.1", "")
	return mainRepo
}

// setupMergeBackPipelineLocal: root-bump seed + wrk wt ahead (no origin).
// Caller sets Args. RepoDir is the linked worktree.
func setupMergeBackPipelineLocal(t *testing.T, req *Request) {
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

// setupMergeBackPipelineWithOrigin: local fixture + bare origin upstream on main.
func setupMergeBackPipelineWithOrigin(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)
	mainRepo := seedMainWithRootBumpTag(t, req)

	bare := setupMergeBackPipelineBareOrigin(t, req.WorkRoot, "origin")
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

// setupMergeBackPipelineSync: two-worktree fixture + v0.0.1 on main (no origin).
// wtA ahead (feature-work); wtB feature-stays at pre-ahead tip.
func setupMergeBackPipelineSync(t *testing.T, req *Request) {
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

// setupMergeBackPipelineSyncWithOrigin: sync fixture + bare origin.
func setupMergeBackPipelineSyncWithOrigin(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)
	mainRepo := seedMainWithRootBumpTag(t, req)

	bare := setupMergeBackPipelineBareOrigin(t, req.WorkRoot, "origin")
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

func mergeBackPushConfirmLine() string {
	return "pushed main → origin/main\n"
}

func primaryMergeMsg(wtBranch string) string {
	return fmt.Sprintf("merged branch %s into main\n", wtBranch)
}

// assertSourceWorktreeKept: merge-back never removes source wt / branch.
func assertSourceWorktreeKept(t *testing.T, req *Request) {
	t.Helper()
	assertFileExists(t, req.WtDir)
	assertGitFileIsWorktreeLink(t, req.WtDir)
	assertBranchExists(t, req.MainRepo, req.WtBranch)
	assertWorktreeListContains(t, req.MainRepo, req.WtDir)
}

// --- events.jsonl helpers (command stays "merge-back" under composition) ---

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

func assertLastEventCommandMergeBack(t *testing.T, wrkHome string) {
	t.Helper()
	events := readPipelineEvents(t, wrkHome)
	if len(events) == 0 {
		t.Fatal("expected at least one events.jsonl entry")
	}
	ev := events[len(events)-1]
	if ev.Command != "merge-back" {
		t.Fatalf("event command: want %q, got %q (args=%v)", "merge-back", ev.Command, ev.Args)
	}
}

func assertLocalTagAtMainHEAD(t *testing.T, mainRepo, tag string) {
	t.Helper()
	if !tagRefExists(t, mainRepo, tag) {
		t.Fatalf("local tag %s should exist after merge-back --tag-next", tag)
	}
	got := revParseRef(t, mainRepo, tag)
	head := revParseHEAD(t, mainRepo)
	if got != head {
		t.Fatalf("%s should point at main HEAD: tag=%s head=%s", tag, got, head)
	}
}

// --- P7: merge-back + --propagate-tags (mirror done-pipeline helpers; no inheritance across trees) ---

const (
	mbPipelinePropOldTag     = "v0.0.1"
	mbPipelinePropNextTag    = "v0.0.2"
	mbPipelinePropModulePath = "example.com/lib"
	mbPipelinePropAppModule  = "example.com/app"
)

type mbPipelineProjectsJSONEntry struct {
	Path    string `json:"path"`
	AddedAt string `json:"added_at"`
	Source  string `json:"source"`
}

type mbPipelineProjectsJSONFile struct {
	Version  int                           `json:"version"`
	Projects []mbPipelineProjectsJSONEntry `json:"projects"`
}

func mbPipelineWriteProjectsJSON(t *testing.T, wrkHome string, paths ...string) {
	t.Helper()
	var projects []mbPipelineProjectsJSONEntry
	for _, p := range paths {
		projects = append(projects, mbPipelineProjectsJSONEntry{
			Path:    p,
			AddedAt: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			Source:  "manual",
		})
	}
	pf := mbPipelineProjectsJSONFile{Version: 1, Projects: projects}
	data, err := json.MarshalIndent(pf, "", "  ")
	if err != nil {
		t.Fatalf("marshal projects.json: %v", err)
	}
	if err := os.MkdirAll(wrkHome, 0o755); err != nil {
		t.Fatalf("mkdir WRK_HOME: %v", err)
	}
	if err := os.WriteFile(filepath.Join(wrkHome, "projects.json"), append(data, '\n'), 0o644); err != nil {
		t.Fatalf("write projects.json: %v", err)
	}
}

func mbPipelineInitGitRepo(t *testing.T, path string) {
	t.Helper()
	mkdirAll(t, path)
	runGitIsolated(t, path, "-c", "init.templateDir=", "init", "-b", "main")
	runGitIsolated(t, path, "config", "user.email", "test@test.com")
	runGitIsolated(t, path, "config", "user.name", "Test")
}

func mbPipelineWriteGoMod(t *testing.T, dir, modulePath string, requires []string) {
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

func mbPipelineWriteConsumerMain(t *testing.T, dir string, importPaths ...string) {
	t.Helper()
	var b strings.Builder
	b.WriteString("package main\n\nimport (\n")
	for _, p := range importPaths {
		fmt.Fprintf(&b, "\t_ %q\n", p)
	}
	b.WriteString(")\n\nfunc main() {}\n")
	writeFile(t, filepath.Join(dir, "main.go"), b.String())
}

func mbPipelineLibGo(version string) string {
	return "package lib\n\nfunc Version() string { return \"" + version + "\" }\n"
}

func mbPipelineReadFile(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return string(data)
}

func mbPipelineSeedFileModuleProxy(t *testing.T, proxyRoot, modulePath, version, srcDir string) {
	t.Helper()
	vDir := filepath.Join(append([]string{proxyRoot}, strings.Split(modulePath, "/")...)...)
	vDir = filepath.Join(vDir, "@v")
	mkdirAll(t, vDir)
	writeFile(t, filepath.Join(vDir, version+".mod"), mbPipelineReadFile(t, filepath.Join(srcDir, "go.mod")))
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
	if err := mbPipelineWriteModuleZip(zipPath, modulePath, version, srcDir); err != nil {
		t.Fatalf("write module zip %s: %v", zipPath, err)
	}
}

func mbPipelineWriteModuleZip(zipPath, modulePath, version, srcDir string) error {
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
			if path != srcDir && filepath.Base(path) == ".git" {
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
		w, err := zw.Create(prefix + filepath.ToSlash(rel))
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

func mbPipelineEnableFileModuleProxy(t *testing.T, req *Request, proxyRoot string) {
	t.Helper()
	abs, err := filepath.Abs(proxyRoot)
	if err != nil {
		t.Fatalf("abs proxy: %v", err)
	}
	req.ExtraEnv = append(req.ExtraEnv,
		"GOPROXY=file://"+abs,
		"GOSUMDB=off",
		"GONOSUMDB=*",
	)
}

func mbPipelineSeedOldModuleProxyFromContent(t *testing.T, proxyRoot, modulePath, version, libGo string) {
	t.Helper()
	tmp := filepath.Join(filepath.Dir(proxyRoot), "seed-old-"+version)
	mkdirAll(t, tmp)
	mbPipelineWriteGoMod(t, tmp, modulePath, nil)
	writeFile(t, filepath.Join(tmp, "lib.go"), libGo)
	mbPipelineSeedFileModuleProxy(t, proxyRoot, modulePath, version, tmp)
}

func mbPipelineRequireGo(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not available")
	}
}

func mbPipelineSnapshotAppBaseline(t *testing.T, req *Request) {
	t.Helper()
	dir := filepath.Join(req.WorkRoot, "_prop_baseline")
	mkdirAll(t, dir)
	writeFile(t, filepath.Join(dir, "app.gomod"), mbPipelineReadFile(t, filepath.Join(req.SecondRepo, "go.mod")))
	writeFile(t, filepath.Join(dir, "app.head"), revParseHEAD(t, req.SecondRepo)+"\n")
}

func mbPipelineReadAppBaseline(t *testing.T, req *Request, name string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(req.WorkRoot, "_prop_baseline", name))
	if err != nil {
		t.Fatalf("read prop baseline %s: %v", name, err)
	}
	return strings.TrimSpace(string(data))
}

// setupMergeBackPipelinePropagateTagNext: same multi-project shape as done twin; wt kept after apply.
func setupMergeBackPipelinePropagateTagNext(t *testing.T, req *Request) {
	t.Helper()
	skipIfNoGit(t)
	mbPipelineRequireGo(t)

	libPath := filepath.Join(req.WorkRoot, "repos", "lib")
	appPath := filepath.Join(req.WorkRoot, "repos", "app")

	mbPipelineInitGitRepo(t, libPath)
	mbPipelineWriteGoMod(t, libPath, mbPipelinePropModulePath, nil)
	writeFile(t, filepath.Join(libPath, "lib.go"), mbPipelineLibGo(mbPipelinePropOldTag))
	runGitIsolated(t, libPath, "add", ".")
	runGitIsolated(t, libPath, "commit", "-m", "init lib "+mbPipelinePropOldTag)
	createLightweightTag(t, libPath, mbPipelinePropOldTag, "")
	libPath = compositionResolvePath(t, libPath)
	req.MainRepo = libPath
	req.DepModulePath = mbPipelinePropModulePath

	mbPipelineInitGitRepo(t, appPath)
	mbPipelineWriteGoMod(t, appPath, mbPipelinePropAppModule, []string{
		mbPipelinePropModulePath + "@" + mbPipelinePropOldTag,
	})
	mbPipelineWriteConsumerMain(t, appPath, mbPipelinePropModulePath)
	runGitIsolated(t, appPath, "add", ".")
	runGitIsolated(t, appPath, "commit", "-m", "init app")
	appPath = compositionResolvePath(t, appPath)
	req.SecondRepo = appPath

	mbPipelineWriteProjectsJSON(t, req.WrkHome, libPath, appPath)

	wtDir := runWrkFrom(t, req, libPath)
	wtDir = compositionResolvePath(t, wtDir)
	req.WtDir = wtDir
	req.WtBranch = branchName("main", wrkDate, 0)

	writeFile(t, filepath.Join(wtDir, "lib.go"), mbPipelineLibGo(mbPipelinePropNextTag))
	runGitIsolated(t, wtDir, "add", "lib.go")
	runGitIsolated(t, wtDir, "commit", "-m", "bump lib for next tag")
	commitAheadOnWorktree(t, wtDir, "feature-work", "ahead of main")

	proxyRoot := filepath.Join(req.WorkRoot, "modproxy")
	mbPipelineSeedFileModuleProxy(t, proxyRoot, mbPipelinePropModulePath, mbPipelinePropNextTag, wtDir)
	mbPipelineSeedOldModuleProxyFromContent(t, proxyRoot, mbPipelinePropModulePath, mbPipelinePropOldTag,
		mbPipelineLibGo(mbPipelinePropOldTag))
	mbPipelineEnableFileModuleProxy(t, req, proxyRoot)

	req.RepoDir = wtDir
	mbPipelineSnapshotAppBaseline(t, req)
}

func mbPropDepsBumpSubject(depModule, version string) string {
	return fmt.Sprintf("chore(deps): bump %s to %s", depModule, version)
}

func mbPropStageApplyStdout(sourceAbs, modulePath, oldTag, nextTag, appBase, appShort, subject string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "source: %s\n", sourceAbs)
	fmt.Fprintf(&b, "  %s  @ %s  (tag %s)\n\n", modulePath, nextTag, nextTag)
	fmt.Fprintf(&b, "updated %s  (project %s)\n", mbPipelinePropAppModule, appBase)
	fmt.Fprintf(&b, "  %s  %s -> %s\n", modulePath, oldTag, nextTag)
	b.WriteString("  go build ./... ok\n")
	fmt.Fprintf(&b, "  committed %s  %s\n\n", appShort, subject)
	b.WriteString("updated 1 module across 1 project\n")
	return b.String()
}

func mbAssertGoModRequireVersion(t *testing.T, goMod, modulePath, version string) {
	t.Helper()
	for _, line := range strings.Split(goMod, "\n") {
		trim := strings.TrimSpace(line)
		fields := strings.Fields(trim)
		if strings.HasPrefix(trim, "require ") && len(fields) >= 3 && fields[1] == modulePath && fields[2] == version {
			return
		}
		if len(fields) >= 2 && fields[0] == modulePath && fields[1] == version {
			return
		}
	}
	t.Fatalf("go.mod missing require %s %s\n%s", modulePath, version, goMod)
}

func mbAssertAppBumpedAndCommitted(t *testing.T, req *Request, wantVersion, wantSubject string) {
	t.Helper()
	app := req.SecondRepo
	gotMod := mbPipelineReadFile(t, filepath.Join(app, "go.mod"))
	mbAssertGoModRequireVersion(t, gotMod, req.DepModulePath, wantVersion)
	before := mbPipelineReadAppBaseline(t, req, "app.head")
	gotHEAD := revParseHEAD(t, app)
	if gotHEAD == before {
		t.Fatalf("app HEAD did not advance after propagate (still %s)", gotHEAD)
	}
	subject := strings.TrimSpace(gitOutputIsolated(t, app, "log", "-1", "--format=%s"))
	if subject != wantSubject {
		t.Fatalf("commit subject want %q got %q", wantSubject, subject)
	}
}

// keep helpers referenced for inheritance compilation
var (
	_ = setupMergeBackPipelineLocal
	_ = setupMergeBackPipelineWithOrigin
	_ = setupMergeBackPipelineSync
	_ = setupMergeBackPipelineSyncWithOrigin
	_ = setupMergeBackPipelinePropagateTagNext
	_ = joinMajorStages
	_ = tagNextRootBumpApplyStdout
	_ = mergeBackPushConfirmLine
	_ = primaryMergeMsg
	_ = assertSourceWorktreeKept
	_ = assertLastEventCommandMergeBack
	_ = assertLocalTagAtMainHEAD
	_ = assertOriginMainEqualsLocalMain
	_ = remoteTagExists
	_ = shortHEAD
	_ = mbPropStageApplyStdout
	_ = mbPropDepsBumpSubject
	_ = mbAssertAppBumpedAndCommitted
	_ = fmt.Sprintf
)
```
