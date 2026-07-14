# Scenario

**Feature**: `wrk --set-config` merges create UX defaults into `$WRK_HOME/config.json`

```
# management mode (no git required)
wrk --set-config --create [--new-window|--new-terminal|--reuse-terminal|--smart-terminal|--open-in-agent|negatives]
  -> merge only flags present into create.* under config.json
  -> empty stdout preferred on successful write; exit 0

wrk --set-config --show
  -> pretty-print effective config.json (or create section)

# mutual exclusion
--set-config + --list | positional create dir | --no-config | other modes → non-zero
```

## Preconditions

- Isolated `{WRK_HOME}` per leaf (root `Setup`).
- Management does not require a git checkout; default cwd is `{WorkRoot}`.
- Config path is always `{WRK_HOME}/config.json`.

## Steps

- Leaves set `Request.Args` for `--set-config` plus section/flags.
- Merge leaves may run wrk twice via helpers and assert final file state.
- Assert exit code, empty or JSON stdout, and `create.*` merge semantics.

## Context

- `--create` selects the `create` section (only section in v1).
- `--new-window` under set-config also persists `terminal.mode=new` (implication).
- Negatives clear/disable that axis (`window` removed/off; `terminal` removed/off; `agent.enabled=false`).
- Unknown top-level keys are preserved.
- Reuses root `Request` / `Run` / `ExtraEnv` harness; no nested `DOCTEST.md`.

```go
import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	if req.RepoDir == "" {
		req.RepoDir = req.WorkRoot
	}
	ensureSetConfigHelpersUsed()
	return nil
}

func setConfigPath(wrkHome string) string {
	return filepath.Join(wrkHome, "config.json")
}

func setConfigArgs(extra ...string) []string {
	args := []string{"--set-config"}
	return append(args, extra...)
}

func writeSetConfigRaw(t *testing.T, wrkHome, content string) {
	t.Helper()
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	writeFile(t, setConfigPath(wrkHome), content)
}

func readSetConfigBytes(t *testing.T, wrkHome string) []byte {
	t.Helper()
	data, err := os.ReadFile(setConfigPath(wrkHome))
	if err != nil {
		t.Fatalf("read config.json: %v", err)
	}
	return data
}

func readSetConfigRoot(t *testing.T, wrkHome string) map[string]json.RawMessage {
	t.Helper()
	var root map[string]json.RawMessage
	if err := json.Unmarshal(readSetConfigBytes(t, wrkHome), &root); err != nil {
		t.Fatalf("parse config.json: %v", err)
	}
	return root
}

func readCreateSection(t *testing.T, wrkHome string) map[string]json.RawMessage {
	t.Helper()
	root := readSetConfigRoot(t, wrkHome)
	createRaw, ok := root["create"]
	if !ok {
		t.Fatal("expected create section in config.json")
	}
	var create map[string]json.RawMessage
	if err := json.Unmarshal(createRaw, &create); err != nil {
		t.Fatalf("parse create: %v", err)
	}
	return create
}

func decodeModeObject(t *testing.T, raw json.RawMessage) string {
	t.Helper()
	var obj struct {
		Mode string `json:"mode"`
	}
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatalf("parse mode object: %v raw=%s", err, string(raw))
	}
	return obj.Mode
}

func decodeAgent(t *testing.T, raw json.RawMessage) (enabled bool, runner, prompt string, args []string) {
	t.Helper()
	var obj struct {
		Enabled         *bool    `json:"enabled"`
		Runner          string   `json:"runner"`
		PromptTemplate  string   `json:"prompt_template"`
		Args            []string `json:"args"`
	}
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatalf("parse agent: %v", err)
	}
	if obj.Enabled == nil {
		t.Fatal("agent.enabled missing")
	}
	return *obj.Enabled, obj.Runner, obj.PromptTemplate, obj.Args
}

func defaultAgentArgs() []string {
	return []string{"--session-id-from-prompt", "--no-submit", "--open"}
}

func assertDefaultAgentOn(t *testing.T, wrkHome string) {
	t.Helper()
	create := readCreateSection(t, wrkHome)
	raw, ok := create["agent"]
	if !ok {
		t.Fatal("expected create.agent")
	}
	enabled, runner, prompt, args := decodeAgent(t, raw)
	if !enabled {
		t.Fatal("agent.enabled want true")
	}
	if runner != "grok-tty" {
		t.Fatalf("agent.runner: want grok-tty, got %q", runner)
	}
	if prompt != "/brainstorm ${task}" {
		t.Fatalf("agent.prompt_template: want /brainstorm ${task}, got %q", prompt)
	}
	wantArgs := defaultAgentArgs()
	if len(args) != len(wantArgs) {
		t.Fatalf("agent.args: want %v, got %v", wantArgs, args)
	}
	for i := range wantArgs {
		if args[i] != wantArgs[i] {
			t.Fatalf("agent.args[%d]: want %q, got %q", i, wantArgs[i], args[i])
		}
	}
}

func assertWindowModeNew(t *testing.T, wrkHome string) {
	t.Helper()
	create := readCreateSection(t, wrkHome)
	raw, ok := create["window"]
	if !ok {
		t.Fatal("expected create.window")
	}
	if mode := decodeModeObject(t, raw); mode != "new" {
		t.Fatalf("window.mode: want new, got %q", mode)
	}
}

func assertTerminalMode(t *testing.T, wrkHome, want string) {
	t.Helper()
	create := readCreateSection(t, wrkHome)
	raw, ok := create["terminal"]
	if !ok {
		t.Fatal("expected create.terminal")
	}
	if mode := decodeModeObject(t, raw); mode != want {
		t.Fatalf("terminal.mode: want %q, got %q", want, mode)
	}
}

func assertWindowAbsentOrOff(t *testing.T, wrkHome string) {
	t.Helper()
	create := readCreateSection(t, wrkHome)
	raw, ok := create["window"]
	if !ok {
		return
	}
	// tolerate explicit off/empty mode if implementer prefers
	var obj map[string]interface{}
	if err := json.Unmarshal(raw, &obj); err != nil {
		t.Fatalf("parse window: %v", err)
	}
	if mode, _ := obj["mode"].(string); mode == "new" {
		t.Fatalf("window should be cleared/off, still mode=new: %s", string(raw))
	}
}

func assertAgentEnabled(t *testing.T, wrkHome string, want bool) {
	t.Helper()
	create := readCreateSection(t, wrkHome)
	raw, ok := create["agent"]
	if !ok {
		if want {
			t.Fatal("expected create.agent")
		}
		return
	}
	enabled, _, _, _ := decodeAgent(t, raw)
	if enabled != want {
		t.Fatalf("agent.enabled: want %v, got %v", want, enabled)
	}
}

func assertEmptyStdout(t *testing.T, stdout string) {
	t.Helper()
	if stdout != "" {
		t.Fatalf("stdout should be empty, got %q", stdout)
	}
}

// runWrkSetConfig runs wrk once with isolated WRK_HOME (same leaf) for multi-step merge leaves.
func runWrkSetConfig(t *testing.T, req *Request, args ...string) *Response {
	t.Helper()
	bin := getWrkBin(t)
	cmd := exec.Command(bin, args...)
	cmd.Dir = req.RepoDir
	if cmd.Dir == "" {
		cmd.Dir = req.WorkRoot
	}
	cmd.Env = append(os.Environ(), "WRK_HOME="+req.WrkHome, "WRK_DATE="+wrkDate)
	var stdout, stderr strings.Builder
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	exit := 0
	if err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			exit = ee.ExitCode()
		} else {
			t.Fatalf("wrk %v: %v", args, err)
		}
	}
	return &Response{Stdout: stdout.String(), Stderr: stderr.String(), ExitCode: exit}
}

func ensureSetConfigHelpersUsed() {
	_ = setConfigPath
	_ = setConfigArgs
	_ = writeSetConfigRaw
	_ = readSetConfigBytes
	_ = readSetConfigRoot
	_ = readCreateSection
	_ = decodeModeObject
	_ = decodeAgent
	_ = defaultAgentArgs
	_ = assertDefaultAgentOn
	_ = assertWindowModeNew
	_ = assertTerminalMode
	_ = assertWindowAbsentOrOff
	_ = assertAgentEnabled
	_ = assertEmptyStdout
	_ = runWrkSetConfig
	_ = fmt.Sprintf
}
```
