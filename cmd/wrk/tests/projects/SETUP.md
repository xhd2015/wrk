# Scenario

**Feature**: wrk project persistence and event logging

```
# every wrk invocation may auto-record the resolved main repo
wrk [dir] [mode] -> resolve work dir -> auto-record main repo (if git)

# persistent storage under WRK_HOME
WRK_HOME -> projects.json (deduped main repos) + events.jsonl (append-only log)

# standalone project modes
wrk --projects -> sorted detailed status blocks (one per recorded main repo)
wrk --add <dir> -> manual record + stdout main repo path
wrk --rm <dir> -> delete recorded main repo + stdout path (or empty if idempotent)
```

## Preconditions

- Git must be available.
- Each test uses isolated `WRK_HOME` at `{WorkRoot}/.wrk`.
- Auto-record and event logging apply to every `wrk` invocation (success or failure).

## Steps

- Descendant scenarios choose invocation form (no-dir cwd, `<dir>` arg, mode flags).
- Assertions inspect `{WRK_HOME}/projects.json` and `{WRK_HOME}/events.jsonl`.

## Context

- `projects.json` stores absolute main-repo paths with `source` (`auto` or `manual`) and ISO-8601 `added_at`.
- `events.jsonl` appends one JSON object per line with `ts`, `command`, `work_dir`, `main_repo`, `args`, `exit_code`.
- Re-adding an existing project path is idempotent (no duplicate entries; first `source` wins).

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
	"github.com/xhd2015/doctest/session"
)

type projectsFile struct {
	Version  int            `json:"version"`
	Projects []projectEntry `json:"projects"`
}

type projectEntry struct {
	Path    string `json:"path"`
	AddedAt string `json:"added_at"`
	Source  string `json:"source"`
}

type wrkEvent struct {
	TS       string   `json:"ts"`
	Command  string   `json:"command"`
	WorkDir  string   `json:"work_dir"`
	MainRepo string   `json:"main_repo"`
	Args     []string `json:"args"`
	ExitCode int      `json:"exit_code"`
}

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}

func projectsJSONPath(wrkHome string) string {
	return filepath.Join(wrkHome, "projects.json")
}

func eventsJSONLPath(wrkHome string) string {
	return filepath.Join(wrkHome, "events.jsonl")
}

func readProjectsFile(t *testing.T, wrkHome string) projectsFile {
	t.Helper()
	data, err := os.ReadFile(projectsJSONPath(wrkHome))
	if err != nil {
		if os.IsNotExist(err) {
			return projectsFile{}
		}
		t.Fatalf("read projects.json: %v", err)
	}
	var pf projectsFile
	if err := json.Unmarshal(data, &pf); err != nil {
		t.Fatalf("parse projects.json: %v", err)
	}
	return pf
}

func readEvents(t *testing.T, wrkHome string) []wrkEvent {
	t.Helper()
	data, err := os.ReadFile(eventsJSONLPath(wrkHome))
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		t.Fatalf("read events.jsonl: %v", err)
	}
	var events []wrkEvent
	for _, line := range strings.Split(strings.TrimSpace(string(data)), "\n") {
		if line == "" {
			continue
		}
		var ev wrkEvent
		if err := json.Unmarshal([]byte(line), &ev); err != nil {
			t.Fatalf("parse event line %q: %v", line, err)
		}
		events = append(events, ev)
	}
	return events
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

func projectsStdoutV2(t *testing.T, blocks ...string) string {
	t.Helper()
	return v2StdoutTemplate(joinStdoutBlocks(blocks...))
}

func assertProjectRecorded(t *testing.T, wrkHome, wantPath, wantSource string) {
	t.Helper()
	pf := readProjectsFile(t, wrkHome)
	want := resolvePath(t, wantPath)
	for _, p := range pf.Projects {
		got := resolvePath(t, p.Path)
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

func assertProjectsCount(t *testing.T, wrkHome string, want int) {
	t.Helper()
	pf := readProjectsFile(t, wrkHome)
	if len(pf.Projects) != want {
		t.Fatalf("projects count: want %d, got %d (%+v)", want, len(pf.Projects), pf.Projects)
	}
	if want > 0 && pf.Version != 1 {
		t.Fatalf("projects.json version: want 1, got %d", pf.Version)
	}
}

func assertNoProjectsFile(t *testing.T, wrkHome string) {
	t.Helper()
	assertFileNotExists(t, projectsJSONPath(wrkHome))
}

func assertProjectsSortedStdout(t *testing.T, stdout string, wantPaths []string) {
	t.Helper()
	var wantLines []string
	for _, p := range wantPaths {
		wantLines = append(wantLines, resolvePath(t, p))
	}
	sort.Strings(wantLines)
	want := strings.Join(wantLines, "\n")
	got := strings.TrimSpace(stdout)
	if got != want {
		t.Fatalf("stdout:\nwant:\n%q\ngot:\n%q", want, got)
	}
}

func assertEventsCount(t *testing.T, wrkHome string, want int) {
	t.Helper()
	events := readEvents(t, wrkHome)
	if len(events) != want {
		t.Fatalf("events count: want %d, got %d", want, len(events))
	}
}

func assertLastEvent(t *testing.T, wrkHome, wantCommand string, wantExit int, wantMainRepo, wantWorkDir string, wantArgs []string) {
	t.Helper()
	events := readEvents(t, wrkHome)
	if len(events) == 0 {
		t.Fatal("no events recorded")
	}
	ev := events[len(events)-1]
	if ev.Command != wantCommand {
		t.Fatalf("event command: want %q, got %q", wantCommand, ev.Command)
	}
	if ev.ExitCode != wantExit {
		t.Fatalf("event exit_code: want %d, got %d", wantExit, ev.ExitCode)
	}
	wantMain := ""
	if wantMainRepo != "" {
		wantMain = resolvePath(t, wantMainRepo)
	}
	gotMain := ""
	if ev.MainRepo != "" {
		gotMain = resolvePath(t, ev.MainRepo)
	}
	if gotMain != wantMain {
		t.Fatalf("event main_repo: want %q, got %q", wantMain, gotMain)
	}
	wantWD := resolvePath(t, wantWorkDir)
	gotWD := resolvePath(t, ev.WorkDir)
	if gotWD != wantWD {
		t.Fatalf("event work_dir: want %q, got %q", wantWD, gotWD)
	}
	if wantArgs == nil {
		wantArgs = []string{}
	}
	if ev.Args == nil {
		ev.Args = []string{}
	}
	if fmt.Sprint(ev.Args) != fmt.Sprint(wantArgs) {
		t.Fatalf("event args: want %v, got %v", wantArgs, ev.Args)
	}
	if ev.TS == "" {
		t.Fatal("event missing ts")
	}
	if _, err := time.Parse(time.RFC3339, ev.TS); err != nil {
		t.Fatalf("event ts not RFC3339: %q", ev.TS)
	}
}

func initProjectsRepo(t *testing.T, workRoot, name string) string {
	t.Helper()
	path := filepath.Join(workRoot, name)
	initGitRepoOnMain(t, path)
	return path
}

func setupLinkedWorktree(t *testing.T, mainRepo, wtName, branch string) string {
	t.Helper()
	wtDir := filepath.Join(filepath.Dir(mainRepo), wtName)
	runGitIsolated(t, mainRepo, "worktree", "add", "-b", branch, wtDir)
	return wtDir
}

func ensureProjectsHelpersUsed() {
	_ = projectsJSONPath
	_ = eventsJSONLPath
	_ = readProjectsFile
	_ = readEvents
	_ = resolvePath
	_ = projectsStdoutV2
	_ = assertProjectRecorded
	_ = assertProjectsCount
	_ = assertNoProjectsFile
	_ = assertProjectsSortedStdout
	_ = assertEventsCount
	_ = assertLastEvent
	_ = initProjectsRepo
	_ = setupLinkedWorktree
}
```