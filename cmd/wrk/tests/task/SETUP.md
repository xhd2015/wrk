# Scenario

**Feature**: wrk --task and wrk --set-task for worktree task descriptions

```
# spawn with --task appends slug to dir/branch names
wrk --task "fix login" -> git worktree add -> stdout path + .wrk-task

# --set-task inside worktree renames via git worktree move
wrk --set-task "new desc" -> parse branch -> compute new names -> git worktree move
```

## Preconditions

- Git must be available.
- The wrk binary is built once per test session.

## Steps

- Spawn tests create a git repo and run `wrk --task <desc>` from it.
- Set-task tests run `wrk --set-task <desc>` from inside a linked worktree.
- `WRK_HOME` is isolated per test at `{WorkRoot}/.wrk`.
- `WRK_DATE` is fixed to `2026-06-30`.

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type wrkEvent struct {
	TS       string   `json:"ts"`
	Command  string   `json:"command"`
	WorkDir  string   `json:"work_dir"`
	MainRepo string   `json:"main_repo"`
	Args     []string `json:"args"`
	ExitCode int      `json:"exit_code"`
}

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}

func eventsJSONLPath(wrkHome string) string {
	return filepath.Join(wrkHome, "events.jsonl")
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

func ensureTaskHelpersUsed() {
	_ = eventsJSONLPath
	_ = readEvents
	_ = resolvePath
	_ = assertLastEvent
}
```
