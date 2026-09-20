package wrkcli

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

// readConfigRaw reads $WRK_HOME/config.json for raw JSON assertions.
func readConfigRaw(t *testing.T, home string) string {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(home, "config.json"))
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

// configAgentMap decodes the create.agent object from config.json.
func configAgentMap(t *testing.T, home string) map[string]interface{} {
	t.Helper()
	var root map[string]interface{}
	if err := json.Unmarshal([]byte(readConfigRaw(t, home)), &root); err != nil {
		t.Fatal(err)
	}
	create, _ := root["create"].(map[string]interface{})
	if create == nil {
		t.Fatalf("config.json missing create section: %s", readConfigRaw(t, home))
	}
	agent, _ := create["agent"].(map[string]interface{})
	if agent == nil {
		t.Fatalf("config.json missing create.agent: %s", readConfigRaw(t, home))
	}
	return agent
}

// TestCaptureAgentRunnerEqualsFormSetConfig runs the user-reported command with
// equals-form flags and asserts both agent keys land in config.json.
func TestCaptureAgentRunnerEqualsFormSetConfig(t *testing.T) {
	home := t.TempDir()
	res := Capture(CaptureOpts{Args: []string{"--set-config", "--create", "--agent-runner=dsh-web", "--browser=brave"}, WrkHome: home})
	if res.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%q", res.ExitCode, res.Stderr)
	}
	data := readConfigRaw(t, home)
	if !strings.Contains(data, `"runner": "dsh-web"`) || !strings.Contains(data, `"browser": "brave"`) {
		t.Fatalf("config.json missing runner/browser after equals-form write: %s", data)
	}
	agent := configAgentMap(t, home)
	want := map[string]interface{}{"runner": "dsh-web", "browser": "brave"}
	if !reflect.DeepEqual(agent, want) {
		t.Fatalf("create.agent=%v want %v", agent, want)
	}
}

// TestCaptureAgentRunnerSpaceFormMatchesEquals asserts space form and equals
// form produce byte-identical config.json.
func TestCaptureAgentRunnerSpaceFormMatchesEquals(t *testing.T) {
	equalsHome := t.TempDir()
	res := Capture(CaptureOpts{Args: []string{"--set-config", "--create", "--agent-runner=dsh-web", "--browser=brave"}, WrkHome: equalsHome})
	if res.ExitCode != 0 {
		t.Fatalf("equals exit=%d stderr=%q", res.ExitCode, res.Stderr)
	}
	spaceHome := t.TempDir()
	res = Capture(CaptureOpts{Args: []string{"--set-config", "--create", "--agent-runner", "dsh-web", "--browser", "brave"}, WrkHome: spaceHome})
	if res.ExitCode != 0 {
		t.Fatalf("space exit=%d stderr=%q", res.ExitCode, res.Stderr)
	}
	equalsData := readConfigRaw(t, equalsHome)
	spaceData := readConfigRaw(t, spaceHome)
	if equalsData != spaceData {
		t.Fatalf("space form config differs from equals form:\nequals: %s\nspace:  %s", equalsData, spaceData)
	}
}

// TestCaptureAgentRunnerCanonicalizesRunner asserts compact runner names are
// stored in canonical form.
func TestCaptureAgentRunnerCanonicalizesRunner(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"codex", "codex-tty"},
		{"grok", "grok-tty"},
		{"codex-tty", "codex-tty"},
		{"grok-tty", "grok-tty"},
		{"dsh-web", "dsh-web"},
	} {
		t.Run(tc.in, func(t *testing.T) {
			home := t.TempDir()
			res := Capture(CaptureOpts{Args: []string{"--set-config", "--create", "--agent-runner", tc.in}, WrkHome: home})
			if res.ExitCode != 0 {
				t.Fatalf("exit=%d stderr=%q", res.ExitCode, res.Stderr)
			}
			if !strings.Contains(readConfigRaw(t, home), `"runner": "`+tc.want+`"`) {
				t.Fatalf("config.json runner=%q want %q: %s", tc.in, tc.want, readConfigRaw(t, home))
			}
		})
	}
}

// TestCaptureAgentRunnerRejectsUnsupportedRunner asserts fail-loud validation
// at set time with normalizeCreateAgentRunner's message and no config write.
func TestCaptureAgentRunnerRejectsUnsupportedRunner(t *testing.T) {
	home := t.TempDir()
	res := Capture(CaptureOpts{Args: []string{"--set-config", "--create", "--agent-runner", "bogus"}, WrkHome: home})
	wantErr := `wrk: unsupported create agent runner "bogus" (want codex, codex-tty, grok, grok-tty, or dsh-web)`
	if res.ExitCode == 0 || !strings.Contains(res.Stderr, wantErr) {
		t.Fatalf("exit=%d stderr=%q, want %q", res.ExitCode, res.Stderr, wantErr)
	}
	if _, err := os.Stat(filepath.Join(home, "config.json")); !os.IsNotExist(err) {
		t.Fatalf("config.json must not be written on rejection, stat err=%v", err)
	}
	if _, err := parseSetConfigArgs([]string{"--set-config", "--create", "--agent-runner", "bogus"}); err == nil || err.Error() != wantErr {
		t.Fatalf("parse err=%v, want %q", err, wantErr)
	}
}

// TestCaptureAgentRunnerRequiresValue asserts missing, empty, and equals-empty
// values fail with the requires-a-value error; other flags keep their
// unrecognized-flag behavior.
func TestCaptureAgentRunnerRequiresValue(t *testing.T) {
	for _, args := range [][]string{
		{"--set-config", "--create", "--agent-runner"},
		{"--set-config", "--create", "--agent-runner", ""},
		{"--set-config", "--create", "--agent-runner="},
	} {
		if _, err := parseSetConfigArgs(args); err == nil || !strings.Contains(err.Error(), "--agent-runner requires a value") {
			t.Fatalf("args=%q err=%v, want requires-a-value error", args, err)
		}
	}
	home := t.TempDir()
	res := Capture(CaptureOpts{Args: []string{"--set-config", "--create", "--agent-runner="}, WrkHome: home})
	if res.ExitCode == 0 || !strings.Contains(res.Stderr, "--agent-runner requires a value") {
		t.Fatalf("exit=%d stderr=%q, want requires-a-value error", res.ExitCode, res.Stderr)
	}
	// --browser= (equals-empty) errors the same way.
	if _, err := parseSetConfigArgs([]string{"--set-config", "--create", "--browser="}); err == nil || !strings.Contains(err.Error(), "--browser requires a value") {
		t.Fatalf("browser= err=%v, want requires-a-value error", err)
	}
	// Bool / unknown flags are not split: equals form stays unrecognized.
	for _, arg := range []string{"--new-window=x", "--open-in-agent=x", "--agent-runner-binary=commandcode"} {
		_, err := parseSetConfigArgs([]string{"--set-config", "--create", arg})
		want := "wrk: unrecognized flag: " + arg
		if err == nil || err.Error() != want {
			t.Fatalf("arg=%q err=%v, want %q", arg, err, want)
		}
	}
}

// TestCaptureAgentRunnerSetConfigMergeOnly asserts the runner write merges into
// an existing agent map and creates a minimal one when absent.
func TestCaptureAgentRunnerSetConfigMergeOnly(t *testing.T) {
	t.Run("preserves existing agent keys", func(t *testing.T) {
		home := t.TempDir()
		seed := `{"version":1,"create":{"agent":{"enabled":true,"prompt_template":"/brainstorm ${task}","args":["--open"],"browser":"brave"}}}`
		if err := os.WriteFile(filepath.Join(home, "config.json"), []byte(seed), 0600); err != nil {
			t.Fatal(err)
		}
		res := Capture(CaptureOpts{Args: []string{"--set-config", "--create", "--agent-runner", "dsh-web"}, WrkHome: home})
		if res.ExitCode != 0 {
			t.Fatalf("exit=%d stderr=%q", res.ExitCode, res.Stderr)
		}
		cfg, err := loadConfig(home)
		if err != nil {
			t.Fatal(err)
		}
		a := cfg.Create.Agent
		if a == nil || a.Runner != "dsh-web" || a.Enabled == nil || !*a.Enabled ||
			a.PromptTemplate != "/brainstorm ${task}" || a.Browser != "brave" ||
			!reflect.DeepEqual(a.Args, []string{"--open"}) {
			t.Fatalf("create.agent not preserved: %+v", a)
		}
	})
	t.Run("creates agent with only runner", func(t *testing.T) {
		home := t.TempDir()
		seed := `{"version":1,"create":{"window":{"mode":"new"}}}`
		if err := os.WriteFile(filepath.Join(home, "config.json"), []byte(seed), 0600); err != nil {
			t.Fatal(err)
		}
		res := Capture(CaptureOpts{Args: []string{"--set-config", "--create", "--agent-runner", "dsh-web"}, WrkHome: home})
		if res.ExitCode != 0 {
			t.Fatalf("exit=%d stderr=%q", res.ExitCode, res.Stderr)
		}
		agent := configAgentMap(t, home)
		want := map[string]interface{}{"runner": "dsh-web"}
		if !reflect.DeepEqual(agent, want) {
			t.Fatalf("create.agent=%v want %v", agent, want)
		}
		if strings.Contains(readConfigRaw(t, home), "enabled") {
			t.Fatalf("create.agent.enabled must stay absent: %s", readConfigRaw(t, home))
		}
	})
	t.Run("creates agent with only runner and browser", func(t *testing.T) {
		home := t.TempDir()
		res := Capture(CaptureOpts{Args: []string{"--set-config", "--create", "--agent-runner", "dsh-web", "--browser", "brave"}, WrkHome: home})
		if res.ExitCode != 0 {
			t.Fatalf("exit=%d stderr=%q", res.ExitCode, res.Stderr)
		}
		agent := configAgentMap(t, home)
		want := map[string]interface{}{"runner": "dsh-web", "browser": "brave"}
		if !reflect.DeepEqual(agent, want) {
			t.Fatalf("create.agent=%v want %v", agent, want)
		}
		if strings.Contains(readConfigRaw(t, home), "enabled") {
			t.Fatalf("create.agent.enabled must stay absent: %s", readConfigRaw(t, home))
		}
	})
}

// TestCaptureAgentRunnerWithOpenInAgent asserts --open-in-agent defaults resolve
// against the overridden runner instead of the fallback default runner.
func TestCaptureAgentRunnerWithOpenInAgent(t *testing.T) {
	home := t.TempDir()
	res := Capture(CaptureOpts{Args: []string{"--set-config", "--create", "--open-in-agent", "--agent-runner", "dsh-web"}, WrkHome: home})
	if res.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%q", res.ExitCode, res.Stderr)
	}
	cfg, err := loadConfig(home)
	if err != nil {
		t.Fatal(err)
	}
	a := cfg.Create.Agent
	if a == nil || a.Runner != "dsh-web" || a.Enabled == nil || !*a.Enabled {
		t.Fatalf("create.agent=%+v, want enabled with dsh-web runner", a)
	}
	if !reflect.DeepEqual(a.Args, defaultAgentArgs("dsh-web")) {
		t.Fatalf("args=%v want dsh-web defaults %v", a.Args, defaultAgentArgs("dsh-web"))
	}
}

// TestCaptureAgentRunnerNoBrowserKeepsRunner asserts --no-browser clears only
// the browser key and never the runner.
func TestCaptureAgentRunnerNoBrowserKeepsRunner(t *testing.T) {
	t.Run("no-browser after agent-runner write", func(t *testing.T) {
		home := t.TempDir()
		res := Capture(CaptureOpts{Args: []string{"--set-config", "--create", "--agent-runner", "dsh-web", "--browser", "brave"}, WrkHome: home})
		if res.ExitCode != 0 {
			t.Fatalf("write exit=%d stderr=%q", res.ExitCode, res.Stderr)
		}
		res = Capture(CaptureOpts{Args: []string{"--set-config", "--create", "--no-browser"}, WrkHome: home})
		if res.ExitCode != 0 {
			t.Fatalf("clear exit=%d stderr=%q", res.ExitCode, res.Stderr)
		}
		data := readConfigRaw(t, home)
		if !strings.Contains(data, `"runner": "dsh-web"`) {
			t.Fatalf("--no-browser must keep runner: %s", data)
		}
		if strings.Contains(data, "browser") {
			t.Fatalf("--no-browser must remove browser: %s", data)
		}
		agent := configAgentMap(t, home)
		want := map[string]interface{}{"runner": "dsh-web"}
		if !reflect.DeepEqual(agent, want) {
			t.Fatalf("create.agent=%v want %v", agent, want)
		}
	})
	t.Run("no-browser alone keeps seeded runner", func(t *testing.T) {
		home := t.TempDir()
		seed := `{"version":1,"create":{"agent":{"enabled":true,"runner":"dsh-web","browser":"brave"}}}`
		if err := os.WriteFile(filepath.Join(home, "config.json"), []byte(seed), 0600); err != nil {
			t.Fatal(err)
		}
		res := Capture(CaptureOpts{Args: []string{"--set-config", "--create", "--no-browser"}, WrkHome: home})
		if res.ExitCode != 0 {
			t.Fatalf("exit=%d stderr=%q", res.ExitCode, res.Stderr)
		}
		data := readConfigRaw(t, home)
		if !strings.Contains(data, `"runner": "dsh-web"`) || strings.Contains(data, "browser") {
			t.Fatalf("--no-browser must keep runner and remove browser: %s", data)
		}
	})
}

// TestSetConfigAgentRunnerCreateUsage asserts the create usage documents the
// new flag and example.
func TestSetConfigAgentRunnerCreateUsage(t *testing.T) {
	usage := setConfigCreateUsage()
	if !strings.Contains(usage, "--agent-runner NAME") {
		t.Fatalf("set-config create usage missing --agent-runner NAME:\n%s", usage)
	}
	if !strings.Contains(usage, "create.agent.runner") {
		t.Fatalf("set-config create usage missing create.agent.runner:\n%s", usage)
	}
	if !strings.Contains(usage, "wrk --set-config --create --agent-runner dsh-web --browser brave") {
		t.Fatalf("set-config create usage missing agent-runner example:\n%s", usage)
	}
}
