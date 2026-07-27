# Scenario

**Feature**: wrk --projects performance instrumentation and budgets

```
# WRK_PROJECTS_PERF_LOG=<path> -> JSONL latency events (no stdout/stderr change)
wrk --projects + perf log -> run_start/project_start/phase/worktree_status/run_end

# parallel gather target (post-fix)
12 linked worktrees -> worktree_status_all < 100ms, run_end < 200ms
```

## Preconditions

- Git must be available.
- Tests use isolated `WRK_HOME` at `{WorkRoot}/.wrk`.
- `WRK_PROJECTS_PERF_LOG` is opt-in; unset env → zero perf overhead, no log file.

## Steps

- Descendants build a main repo with N linked worktrees, record via `wrk --add`, set `req.ProjectsPerfLog`, run `wrk --projects`.
- Assertions read the JSONL perf log (not stdout fragments).

## Context

Perf event schema (one JSON object per line):

| event | meaning |
|-------|---------|
| `run_start` / `run_end` | whole `--projects` invocation (`total_ms` on end) |
| `project_start` / `project_end` | per recorded main repo (`total_ms` on end) |
| `phase` | timed step (`phase`, `duration_ms`) |
| `worktree_status` | per linked worktree porcelain check |
| `phase_total` | aggregate (`phase=worktree_status_all`, `count`, `duration_ms`) |

Phases today: `main_branch`, `main_commit_short`, `main_commit_subject`, `main_status`, `remote`, `list_linked_skip`, `list_linked_summary`, `worktree_summary`.

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensurePerfProfileHelpersUsed()
	return nil
}

type projectsPerfEvent struct {
	Event      string `json:"event"`
	Project    string `json:"project"`
	Phase      string `json:"phase"`
	Worktree   string `json:"worktree"`
	DurationMS int64  `json:"duration_ms"`
	Count      int    `json:"count"`
	TotalMS    int64  `json:"total_ms"`
}

func perfLogPath(req *Request) string {
	if req.ProjectsPerfLog != "" {
		return req.ProjectsPerfLog
	}
	return filepath.Join(req.WorkRoot, "perf.jsonl")
}

func initPerfProfileRepo(t *testing.T, path, subject string) {
	t.Helper()
	mkdirAll(t, path)
	runGitIsolated(t, path, "-c", "init.templateDir=", "init", "-b", "main")
	runGitIsolated(t, path, "config", "user.email", "test@test.com")
	runGitIsolated(t, path, "config", "user.name", "Test")
	writeFile(t, filepath.Join(path, "README.md"), "# "+filepath.Base(path)+"\n")
	runGitIsolated(t, path, "add", "README.md")
	runGitIsolated(t, path, "commit", "-m", subject)
}

func setupBareOrigin(t *testing.T, workRoot, name string) string {
	t.Helper()
	bare := filepath.Join(workRoot, name+".git")
	runGitIsolated(t, workRoot, "-c", "init.templateDir=", "init", "--bare", "-b", "main", bare)
	return bare
}

func setupTrackedMainRepo(t *testing.T, workRoot, name, originBare, subject string) string {
	t.Helper()
	repo := filepath.Join(workRoot, name)
	initPerfProfileRepo(t, repo, subject)
	runGitIsolated(t, repo, "remote", "add", "origin", originBare)
	runGitIsolated(t, repo, "push", "-u", "origin", "main")
	return repo
}

func addLinkedWorktreeForProject(t *testing.T, mainRepo, relDir, branch string) string {
	t.Helper()
	wtDir := filepath.Join(mainRepo, filepath.FromSlash(relDir))
	runGitIsolated(t, mainRepo, "worktree", "add", "-b", branch, wtDir)
	return wtDir
}

func seedPerfProfileProject(t *testing.T, req *Request, repoPath string) {
	t.Helper()
	pf := projectsFile{
		Version: 1,
		Projects: []projectEntry{{
			Path:    resolvePath(t, repoPath),
			AddedAt: "2026-01-01T00:00:00Z",
			Source:  "manual",
		}},
	}
	data, err := json.Marshal(pf)
	if err != nil {
		t.Fatalf("marshal projects.json: %v", err)
	}
	if err := os.WriteFile(projectsJSONPath(req.WrkHome), data, 0o644); err != nil {
		t.Fatalf("write projects.json: %v", err)
	}
}

func setupPerfProfileRepo(t *testing.T, req *Request, name string, worktreeCount int) string {
	t.Helper()
	origin := setupBareOrigin(t, req.WorkRoot, name+"-origin")
	repo := setupTrackedMainRepo(t, req.WorkRoot, name, origin, name+" perf profile")
	for i := 1; i <= worktreeCount; i++ {
		branch := fmt.Sprintf("wt-%d", i)
		addLinkedWorktreeForProject(t, repo, branch, branch)
	}
	seedPerfProfileProject(t, req, repo)
	req.MainRepo = repo
	req.ProjectsPerfLog = perfLogPath(req)
	req.Args = []string{"--projects"}
	req.RepoDir = req.WorkRoot
	return repo
}

func readProjectsPerfLog(t *testing.T, path string) []projectsPerfEvent {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read perf log %q: %v", path, err)
	}
	var events []projectsPerfEvent
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var ev projectsPerfEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			t.Fatalf("parse perf line %q: %v", line, err)
		}
		events = append(events, ev)
	}
	return events
}

func perfPhaseMS(events []projectsPerfEvent, phase string) int64 {
	var total int64
	for _, ev := range events {
		if ev.Event == "phase" && ev.Phase == phase {
			total += ev.DurationMS
		}
	}
	return total
}

func perfPhaseTotalMS(events []projectsPerfEvent, phase string) (int64, int) {
	for _, ev := range events {
		if ev.Event == "phase_total" && ev.Phase == phase {
			return ev.DurationMS, ev.Count
		}
	}
	return 0, 0
}

func perfRunEndMS(events []projectsPerfEvent) int64 {
	for _, ev := range events {
		if ev.Event == "run_end" {
			return ev.TotalMS
		}
	}
	return 0
}

func perfWorktreeStatusCount(events []projectsPerfEvent) int {
	n := 0
	for _, ev := range events {
		if ev.Event == "worktree_status" {
			n++
		}
	}
	return n
}

func perfListLinkedPhaseCount(events []projectsPerfEvent) int {
	n := 0
	for _, ev := range events {
		if ev.Event == "phase" && strings.HasPrefix(ev.Phase, "list_linked") {
			n++
		}
	}
	return n
}

func ensurePerfProfileHelpersUsed() {
	_ = perfLogPath
	_ = setupPerfProfileRepo
	_ = readProjectsPerfLog
	_ = seedPerfProfileProject
	_ = perfPhaseMS
	_ = perfPhaseTotalMS
	_ = perfRunEndMS
	_ = perfWorktreeStatusCount
	_ = perfListLinkedPhaseCount
	_ = initPerfProfileRepo
	_ = setupBareOrigin
	_ = setupTrackedMainRepo
	_ = addLinkedWorktreeForProject
}
```