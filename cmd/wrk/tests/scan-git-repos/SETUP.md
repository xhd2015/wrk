# Scenario

**Feature**: wrk --scan-git-repos discovers main git repos and records them in projects.json

```
# standalone mode: scan roots for main repos, record with source=scan
wrk --scan-git-repos [ROOT...] [--no-cache]
  -> scan_repo.Scan(Roots, NoCache)
  -> filter RepoTypeMain only
  -> storage.RecordProject(wrkHome, path, source="scan") for each main
  -> stdout: newly recorded absolute main paths (one per line)

# already recorded
wrk --scan-git-repos ROOT (2nd run)
  -> exit 0; no duplicate projects.json entries; no newly-added stdout lines

# cache flag
wrk --scan-git-repos --no-cache ROOT  -> still discovers + records (NoCache to scan_repo)
wrk --no-cache (no --scan-git-repos)  -> non-zero; only valid with --scan-git-repos

# mutual exclusion / help
wrk --scan-git-repos --projects  -> non-zero; mutually exclusive
wrk -h  -> documents --scan-git-repos and --no-cache
```

## Preconditions

- Git must be available for discovery leaves.
- Each test isolates `WRK_HOME` at `{WorkRoot}/.wrk` (root Setup).
- Process cwd is `{WorkRoot}` (non-git) so auto-record does not pollute `projects.json`.
- Explicit scan roots under `{WorkRoot}` — do not rely on `$HOME/Projects`.

## Steps

1. Root Setup creates `WorkRoot` / `WrkHome`.
2. This Setup sets `RepoDir = WorkRoot` and registers scan helpers.
3. Descendants build git fixtures under a scan root and set `req.Args`.

## Context

- `projects.json` entries use `source: "scan"` for paths first seen via `--scan-git-repos`.
- Re-recording is idempotent (first `source` wins; no duplicate paths).
- Linked worktrees may appear under the scan root; only `RepoTypeMain` paths are recorded.
- Stdout prefers **newly** recorded absolute main paths only (one per line, trailing `\n`); already-known stays silent on stdout.

```go
import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
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

func Setup(t *testing.T, req *Request) error {
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

// seedScanProject writes a projects.json entry as if a prior --scan-git-repos
// had recorded path with source "scan". Used for idempotent leaves so Setup
// does not depend on the feature under test.
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
}
```
