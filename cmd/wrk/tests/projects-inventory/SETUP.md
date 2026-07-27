# Scenario

**Feature**: pure inventory graph + source release helpers for registered projects

```
# WRK_HOME projects.json -> BuildInventory -> modules + ownership + cross/intra edges
# source main repo tags -> ResolveSourceReleases -> ModulePath/Tag/Version (+ Missing)

storage.ListProjects(wrkHome)
  -> scan each existing project for go.mod modules
  -> Inventory{Projects, SkippedPaths}
  -> CrossEdges() / IntraEdges() / FindOwner(modulePath)

sourceMain
  -> ResolveSourceReleases
  -> Releases[] + Missing[]
```

## Preconditions

- Nested root: **no inheritance** from `cmd/wrk/tests` monotree (`DOCTEST.md` firewall).
- Package `github.com/xhd2015/wrk/wrkcli` is the production import target.
- Git and Go toolchain available on PATH (fixtures create real repos + go.mod).
- Helpers use `github.com/xhd2015/gitops/git/git_isolated` (hook-free).
- No wrk binary build; no CLI flags.

## Steps

1. Root Setup creates isolated `WorkRoot` and `WrkHome` (`{WorkRoot}/.wrk`).
2. Grouping/leaf Setup writes git fixtures + `projects.json` (or tags) and sets `Op`.
3. Root `Run` calls `BuildInventory` or `ResolveSourceReleases` per `Op`.
4. Leaf `Assert` checks modules, edges, or releases against `Want*`.

## Context

- **projects.json** schema matches `storage.ProjectsFile` (`version`, `projects[{path,added_at,source}]`).
- Module Dir uses scan convention: `"."` for root module, slash-joined relative dirs for nested.
- Absolute paths in Response are compared after `storage.NormalizePath`.
- Classic TDD: missing API symbols → RED until implementer lands helpers.

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
	"time"

	"github.com/xhd2015/gitops/git/git_isolated"
	"github.com/xhd2015/wrk/wrkcli/storage"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	if _, err := exec.LookPath("go"); err != nil {
		return fmt.Errorf("go not found in PATH: %w", err)
	}
	workRoot, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		return fmt.Errorf("resolve work root: %w", err)
	}
	req.WorkRoot = workRoot
	req.WrkHome = filepath.Join(req.WorkRoot, ".wrk")
	if err := os.MkdirAll(req.WrkHome, 0o755); err != nil {
		return err
	}
	// Defaults for Want slices so Assert never sees nil vs empty confusion.
	if req.WantProjectPaths == nil {
		req.WantProjectPaths = []string{}
	}
	if req.WantModules == nil {
		req.WantModules = []WantModule{}
	}
	if req.WantCrossEdges == nil {
		req.WantCrossEdges = []WantEdge{}
	}
	if req.WantIntraEdges == nil {
		req.WantIntraEdges = []WantEdge{}
	}
	if req.WantSkippedPaths == nil {
		req.WantSkippedPaths = []string{}
	}
	if req.WantReleases == nil {
		req.WantReleases = []WantRelease{}
	}
	if req.WantMissing == nil {
		req.WantMissing = []string{}
	}
	inventoryEnsureHelpersUsed()
	return nil
}

func skipIfNoGit(t *testing.T) {
	t.Helper()
	if _, err := exec.LookPath("git"); err != nil {
		t.Skip("git not available")
	}
}

func mkdirAll(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(path, 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", path, err)
	}
}

func writeFile(t *testing.T, path, content string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir %s: %v", filepath.Dir(path), err)
	}
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}

func runGitIsolated(t *testing.T, dir string, args ...string) {
	t.Helper()
	git_isolated.MustRun(t, dir, args...)
}

func resolvePath(t *testing.T, path string) string {
	t.Helper()
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("abs %s: %v", path, err)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return abs
	}
	return resolved
}

func normPath(p string) string {
	return storage.NormalizePath(p)
}

func normPaths(ps []string) []string {
	if ps == nil {
		return []string{}
	}
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = normPath(p)
	}
	return out
}

type projectsJSONEntry struct {
	Path    string `json:"path"`
	AddedAt string `json:"added_at"`
	Source  string `json:"source"`
}

type projectsJSONFile struct {
	Version  int                 `json:"version"`
	Projects []projectsJSONEntry `json:"projects"`
}

// writeProjectsJSON seeds WRK_HOME/projects.json with the given absolute paths.
// Paths need not exist (used for soft-skip scenarios).
func writeProjectsJSON(t *testing.T, wrkHome string, paths ...string) {
	t.Helper()
	var projects []projectsJSONEntry
	for _, p := range paths {
		projects = append(projects, projectsJSONEntry{
			Path:    p,
			AddedAt: time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339),
			Source:  "manual",
		})
	}
	pf := projectsJSONFile{Version: 1, Projects: projects}
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

// initGitRepo creates an empty git repo on branch main with user config.
func initGitRepo(t *testing.T, path string) {
	t.Helper()
	mkdirAll(t, path)
	runGitIsolated(t, path, "-c", "init.templateDir=", "init", "-b", "main")
	runGitIsolated(t, path, "config", "user.email", "test@test.com")
	runGitIsolated(t, path, "config", "user.name", "Test")
}

// writeGoMod writes a minimal go.mod with optional require lines.
// requires are pairs of modulePath@version strings like "example.com/lib@v1.0.0".
func writeGoMod(t *testing.T, dir, modulePath string, requires ...string) {
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

// initRootAndSubModuleRepo creates a git repo with root module and nested sub/.
// rootModule is e.g. example.com/lib; sub module path is rootModule+"/sub".
// rootRequires are optional root-level requires (path@version).
func initRootAndSubModuleRepo(t *testing.T, path, rootModule string, rootRequires ...string) {
	t.Helper()
	initGitRepo(t, path)
	writeGoMod(t, path, rootModule, rootRequires...)
	writeFile(t, filepath.Join(path, "root.go"), "package rootmod\n")
	subDir := filepath.Join(path, "sub")
	mkdirAll(t, subDir)
	writeGoMod(t, subDir, rootModule+"/sub")
	writeFile(t, filepath.Join(subDir, "sub.go"), "package sub\n")
	runGitIsolated(t, path, "add", ".")
	runGitIsolated(t, path, "commit", "-m", "init root+sub modules")
}

// initSingleModuleRepo creates a git repo with only a root go.mod.
func initSingleModuleRepo(t *testing.T, path, modulePath string, requires ...string) {
	t.Helper()
	initGitRepo(t, path)
	writeGoMod(t, path, modulePath, requires...)
	writeFile(t, filepath.Join(path, "main.go"), "package main\n")
	runGitIsolated(t, path, "add", ".")
	runGitIsolated(t, path, "commit", "-m", "init single module")
}

// tagRepo creates a lightweight git tag at HEAD.
func tagRepo(t *testing.T, path, tag string) {
	t.Helper()
	runGitIsolated(t, path, "tag", tag)
}

func assertErrIsNil(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func assertStringSlicesEqual(t *testing.T, got, want []string, label string) {
	t.Helper()
	g := append([]string(nil), got...)
	w := append([]string(nil), want...)
	sort.Strings(g)
	sort.Strings(w)
	if !reflect.DeepEqual(g, w) {
		t.Fatalf("%s mismatch\n got: %#v\nwant: %#v", label, g, w)
	}
}

func assertNormPathSlicesEqual(t *testing.T, got, want []string, label string) {
	t.Helper()
	g := normPaths(got)
	w := normPaths(want)
	sort.Strings(g)
	sort.Strings(w)
	if !reflect.DeepEqual(g, w) {
		t.Fatalf("%s mismatch\n got: %#v\nwant: %#v", label, g, w)
	}
}

func moduleKey(projectPath, dir, path string) string {
	return normPath(projectPath) + "|" + dir + "|" + path
}

func assertModulesMatch(t *testing.T, got []ModuleSnap, want []WantModule) {
	t.Helper()
	gotMap := map[string]ModuleSnap{}
	for _, m := range got {
		gotMap[moduleKey(m.ProjectPath, m.Dir, m.Path)] = m
	}
	wantMap := map[string]WantModule{}
	for _, m := range want {
		wantMap[moduleKey(m.ProjectPath, m.Dir, m.Path)] = m
	}
	if len(gotMap) != len(wantMap) {
		t.Fatalf("module count got %d want %d\n got: %#v\nwant: %#v",
			len(gotMap), len(wantMap), got, want)
	}
	for k, w := range wantMap {
		g, ok := gotMap[k]
		if !ok {
			t.Fatalf("missing module %s\n got: %#v\nwant: %#v", k, got, want)
		}
		if g.Dir != w.Dir || g.Path != w.Path {
			t.Fatalf("module %s Dir/Path got {%q,%q} want {%q,%q}", k, g.Dir, g.Path, w.Dir, w.Path)
		}
		if normPath(g.ProjectPath) != normPath(w.ProjectPath) {
			t.Fatalf("module %s ProjectPath got %q want %q", k, g.ProjectPath, w.ProjectPath)
		}
	}
}

func edgeKey(e EdgeSnap) string {
	return strings.Join([]string{
		normPath(e.ConsumerProject),
		e.ConsumerModule,
		e.DepPath,
		normPath(e.OwnerProject),
	}, "|")
}

func wantEdgeKey(e WantEdge) string {
	return strings.Join([]string{
		normPath(e.ConsumerProject),
		e.ConsumerModule,
		e.DepPath,
		normPath(e.OwnerProject),
	}, "|")
}

func assertEdgesMatch(t *testing.T, got []EdgeSnap, want []WantEdge, label string) {
	t.Helper()
	gotMap := map[string]EdgeSnap{}
	for _, e := range got {
		gotMap[edgeKey(e)] = e
	}
	wantMap := map[string]WantEdge{}
	for _, e := range want {
		wantMap[wantEdgeKey(e)] = e
	}
	if len(gotMap) != len(wantMap) {
		t.Fatalf("%s count got %d want %d\n got: %#v\nwant: %#v",
			label, len(gotMap), len(wantMap), got, want)
	}
	for k, w := range wantMap {
		g, ok := gotMap[k]
		if !ok {
			t.Fatalf("%s missing edge %s\n got: %#v\nwant: %#v", label, k, got, want)
		}
		if w.DepVersion != "" && g.DepVersion != w.DepVersion {
			t.Fatalf("%s edge %s DepVersion got %q want %q", label, k, g.DepVersion, w.DepVersion)
		}
	}
}

func assertReleasesMatch(t *testing.T, got []ReleaseSnap, want []WantRelease) {
	t.Helper()
	gotMap := map[string]ReleaseSnap{}
	for _, r := range got {
		gotMap[r.ModulePath] = r
	}
	wantMap := map[string]WantRelease{}
	for _, r := range want {
		wantMap[r.ModulePath] = r
	}
	if len(gotMap) != len(wantMap) {
		t.Fatalf("releases count got %d want %d\n got: %#v\nwant: %#v",
			len(gotMap), len(wantMap), got, want)
	}
	for mod, w := range wantMap {
		g, ok := gotMap[mod]
		if !ok {
			t.Fatalf("missing release for module %s\n got: %#v\nwant: %#v", mod, got, want)
		}
		if g.Tag != w.Tag {
			t.Fatalf("module %s Tag got %q want %q", mod, g.Tag, w.Tag)
		}
		if g.Version != w.Version {
			t.Fatalf("module %s Version got %q want %q", mod, g.Version, w.Version)
		}
	}
}

func assertInventorySuccess(t *testing.T, req *Request, resp *Response, err error) {
	t.Helper()
	assertErrIsNil(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	assertNormPathSlicesEqual(t, resp.ProjectPaths, req.WantProjectPaths, "ProjectPaths")
	assertNormPathSlicesEqual(t, resp.SkippedPaths, req.WantSkippedPaths, "SkippedPaths")
	assertModulesMatch(t, resp.Modules, req.WantModules)
	assertEdgesMatch(t, resp.CrossEdges, req.WantCrossEdges, "CrossEdges")
	assertEdgesMatch(t, resp.IntraEdges, req.WantIntraEdges, "IntraEdges")
}

func assertSourceReleasesSuccess(t *testing.T, req *Request, resp *Response, err error) {
	t.Helper()
	assertErrIsNil(t, err)
	if resp == nil {
		t.Fatal("resp is nil")
	}
	assertReleasesMatch(t, resp.Releases, req.WantReleases)
	assertStringSlicesEqual(t, resp.Missing, req.WantMissing, "Missing")
}

func inventoryEnsureHelpersUsed() {
	_ = writeProjectsJSON
	_ = initRootAndSubModuleRepo
	_ = initSingleModuleRepo
	_ = tagRepo
	_ = assertInventorySuccess
	_ = assertSourceReleasesSuccess
}
```
