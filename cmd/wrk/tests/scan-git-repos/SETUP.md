# Scenario

**Feature**: wrk --scan-git-repos always streams every valid git repo found; never mutates projects.json

```
# standalone mode: scan roots; always print valid finds; print-only (no registry)
wrk --scan-git-repos [ROOT...] [--no-cache] [--include-worktrees]
  -> scan_repo.Scan(Roots, NoCache, OnRepo)
  -> default: print RepoTypeMain only; with --include-worktrees also print worktrees
  -> stdout: every valid abs path as discovered (not post-scan sort)
  -> projects.json is never read or written by this mode

# default root when no ROOT... (see defaults/)
wrk --scan-git-repos  -> roots = [$HOME] if home is a directory (not ~/Projects)
wrk --scan-git-repos  (HOME missing/not a dir)  -> non-zero; error about home/~

# pre-seeded projects (always-print / record/idempotent)
wrk --scan-git-repos ROOT (2nd run / pre-seeded projects)
  -> exit 0; still prints valid main path once; projects.json unchanged

# streaming / interrupt (see streaming/ and interrupt/ leaves)
wrk --scan-git-repos ROOT_B ROOT_A  -> discovery-order stdout (CLI root order)
wrk --scan-git-repos --no-cache ROOT_FIRST ROOT_LATER  -> first path before finish
SIGINT mid-scan  -> exit 130; stderr warning: interrupted; cache may keep progress; projects.json untouched

# cache flag
wrk --scan-git-repos --no-cache ROOT  -> still discovers + prints (NoCache); no projects.json write
wrk --no-cache (no --scan-git-repos)  -> non-zero; only valid with --scan-git-repos

# include-worktrees (see include-worktrees/)
wrk --scan-git-repos --include-worktrees ROOT  -> print main + valid worktrees; no registry write
wrk --include-worktrees (no --scan-git-repos)  -> non-zero; only valid with --scan-git-repos

# debug wiring (see debug/ leaves) — implemented P3
wrk --scan-git-repos -v ROOT  -> Options.Debug=true; stderr scan: + mode=
WRK_SCAN_DEBUG=1 wrk --scan-git-repos ROOT  -> Debug=true without -v; scan: present
wrk --scan-git-repos ROOT (no -v, no env)  -> zero scan: markers
# truthy env: 1, true, yes (case-insensitive); product CacheRoot under FakeHome

# mutual exclusion / help
wrk --scan-git-repos --projects  -> non-zero; mutually exclusive
wrk -h  -> documents --scan-git-repos, --no-cache, --include-worktrees, default root ~, print-only

# P5 two-base + filter (see filter-home-subpath/) — FakeHome isolation
wrk --scan-git-repos  -> universe home; product $HOME/.cache/git-repo-scan/home/
wrk --scan-git-repos ~/Projects  -> same home cache files; stdout only under Projects
wrk --scan-git-repos ~/Projects -v  -> stderr cache_base + filter
```

## Preconditions

- Git must be available for discovery leaves.
- Each test isolates `WRK_HOME` at `{WorkRoot}/.wrk` (root Setup).
- Process cwd is `{WorkRoot}` (non-git) so auto-record does not pollute `projects.json`.
- Explicit-root leaves use roots under `{WorkRoot}` — do not rely on `$HOME/Projects`.
- Bare-flag default-root leaves (`defaults/`) isolate `HOME` via `FakeHome` and must not create `Projects`.
- `filter-home-subpath/` uses FakeHome with both `home-main` and `Projects/proj-main`.

## Steps

1. Root Setup creates `WorkRoot` / `WrkHome`.
2. This Setup sets `RepoDir = WorkRoot` and registers scan helpers.
3. Descendants build git fixtures under a scan root and set `req.Args`.

## Context

- **`--scan-git-repos` never mutates `projects.json`** (print-only forever). Registry writes remain `auto` / `manual` only.
- Historical `source: "scan"` entries may still exist in older files; scan no longer creates them.
- Linked worktrees may appear under the scan root; default omits them from stdout; `--include-worktrees` prints valid worktrees.
- Stdout always lists every **valid** discovery (live `.git`, under root filter, type allowed by flags) **as found** in discovery order (one absolute path per line, trailing `\n`); same path at most once per run. Not batch-sorted after the full scan.
- Mid-scan SIGINT/SIGTERM → exit 130, stderr interrupt warning; scan disk cache may keep progress; `projects.json` unchanged.

```go
import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"time"
	"github.com/xhd2015/doctest/session"
)

type scanProjectsFile struct {
	Version  int                `json:"version"`
	Projects []scanProjectEntry `json:"projects"`
}

type scanProjectEntry struct {
	Path    string `json:"path"`
	AddedAt string `json:"added_at"`
	Source  string `json:"source"`
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	// Non-git cwd so auto-record does not add the wrk worktree itself.
	req.RepoDir = req.WorkRoot
	ensureScanGitReposHelpersUsed()
	return nil
}

func scanProjectsJSONPath(wrkHome string) string {
	return filepath.Join(wrkHome, "projects.json")
}

func readScanProjectsFile(t *testing.T, wrkHome string) scanProjectsFile {
	t.Helper()
	data, err := os.ReadFile(scanProjectsJSONPath(wrkHome))
	if err != nil {
		if os.IsNotExist(err) {
			return scanProjectsFile{}
		}
		t.Fatalf("read projects.json: %v", err)
	}
	var pf scanProjectsFile
	if err := json.Unmarshal(data, &pf); err != nil {
		t.Fatalf("parse projects.json: %v", err)
	}
	return pf
}

func resolveScanPath(t *testing.T, path string) string {
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

func assertScanProjectRecorded(t *testing.T, wrkHome, wantPath, wantSource string) {
	t.Helper()
	pf := readScanProjectsFile(t, wrkHome)
	want := resolveScanPath(t, wantPath)
	for _, p := range pf.Projects {
		got := resolveScanPath(t, p.Path)
		if got == want {
			if wantSource != "" && p.Source != wantSource {
				t.Fatalf("project %q source: want %q, got %q", want, wantSource, p.Source)
			}
			if p.AddedAt == "" {
				t.Fatalf("project %q missing added_at", want)
			}
			if _, err := time.Parse(time.RFC3339, p.AddedAt); err != nil {
				t.Fatalf("project %q added_at not RFC3339: %q", want, p.AddedAt)
			}
			return
		}
	}
	t.Fatalf("projects.json should contain %q (source=%q), got %+v", want, wantSource, pf.Projects)
}

func assertScanProjectsCount(t *testing.T, wrkHome string, want int) {
	t.Helper()
	pf := readScanProjectsFile(t, wrkHome)
	if len(pf.Projects) != want {
		t.Fatalf("projects count: want %d, got %d (%+v)", want, len(pf.Projects), pf.Projects)
	}
	if want > 0 && pf.Version != 1 {
		t.Fatalf("projects.json version: want 1, got %d", pf.Version)
	}
}

func assertScanProjectNotRecorded(t *testing.T, wrkHome, path string) {
	t.Helper()
	pf := readScanProjectsFile(t, wrkHome)
	want := resolveScanPath(t, path)
	for _, p := range pf.Projects {
		if resolveScanPath(t, p.Path) == want {
			t.Fatalf("projects.json must not contain %q (got %+v)", want, pf.Projects)
		}
	}
}

func initScanMainRepo(t *testing.T, workRoot, name string) string {
	t.Helper()
	path := filepath.Join(workRoot, name)
	initGitRepoOnMain(t, path)
	return path
}

// setupScanLinkedWorktree adds a linked worktree as a sibling of mainRepo under the same parent.
func setupScanLinkedWorktree(t *testing.T, mainRepo, wtName, branch string) string {
	t.Helper()
	wtDir := filepath.Join(filepath.Dir(mainRepo), wtName)
	runGitIsolated(t, mainRepo, "worktree", "add", "-b", branch, wtDir)
	return wtDir
}

// makeScanRoot creates {workRoot}/scan-root and returns its absolute path.
func makeScanRoot(t *testing.T, workRoot string) string {
	t.Helper()
	root := filepath.Join(workRoot, "scan-root")
	mkdirAll(t, root)
	return root
}

// seedScanProject writes a projects.json entry with source "scan" for tests that
// need a pre-existing registry row (idempotent / known-main). Scan itself no
// longer writes projects.json; this helper only simulates legacy data.
func seedScanProject(t *testing.T, wrkHome, path string) {
	t.Helper()
	if err := os.MkdirAll(wrkHome, 0o755); err != nil {
		t.Fatalf("mkdir wrk home: %v", err)
	}
	pf := scanProjectsFile{
		Version: 1,
		Projects: []scanProjectEntry{{
			Path:    resolveScanPath(t, path),
			AddedAt: time.Now().UTC().Format(time.RFC3339),
			Source:  "scan",
		}},
	}
	data, err := json.Marshal(pf)
	if err != nil {
		t.Fatalf("marshal projects: %v", err)
	}
	if err := os.WriteFile(scanProjectsJSONPath(wrkHome), data, 0o644); err != nil {
		t.Fatalf("write projects.json: %v", err)
	}
}

// countScanStdoutPathLines counts non-empty stdout lines whose resolved path equals wantPath.
func countScanStdoutPathLines(t *testing.T, stdout, wantPath string) int {
	t.Helper()
	want := resolveScanPath(t, wantPath)
	n := 0
	for _, line := range strings.Split(stdout, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if resolveScanPath(t, line) == want {
			n++
		}
	}
	return n
}

func ensureScanGitReposHelpersUsed() {
	_ = scanProjectsJSONPath
	_ = readScanProjectsFile
	_ = resolveScanPath
	_ = assertScanProjectRecorded
	_ = assertScanProjectsCount
	_ = assertScanProjectNotRecorded
	_ = initScanMainRepo
	_ = setupScanLinkedWorktree
	_ = makeScanRoot
	_ = seedScanProject
	_ = countScanStdoutPathLines
}
```
