package wrkcli

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestDSHWebCreateDefaults(t *testing.T) {
	plan, err := resolveCreateUX(t.TempDir(), createUXFlags{
		openInAgent: true,
		agentRunner: stringPtr("dsh-web"),
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	if plan.runner != "dsh-web" || plan.promptTmpl != "/brainstorm ${task}" {
		t.Fatalf("unexpected plan: %+v", plan)
	}
	if !reflect.DeepEqual(plan.agentArgs, []string{"--open", "--no-submit"}) {
		t.Fatalf("args=%q", plan.agentArgs)
	}
	workspace := t.TempDir()
	got, err := buildAgentArgv(workspace, plan, "fix 'quoted' bug")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"agent-run", "run", "--dir", workspace, "--open", "--no-submit", "--agent-runner=dsh-web", "/brainstorm fix 'quoted' bug"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("argv=%q want %q", got, want)
	}
}

func TestDSHWebPreservesTerminalDefaults(t *testing.T) {
	workspace := t.TempDir()
	for _, runner := range []string{"grok-tty", "codex-tty"} {
		plan, err := resolveCreateUX(t.TempDir(), createUXFlags{
			openInAgent: true, agentRunner: stringPtr(runner),
		}, false)
		if err != nil {
			t.Fatal(err)
		}
		want := []string{"--session-id-from-prompt", "--no-submit", "--open", "--color"}
		if !reflect.DeepEqual(plan.agentArgs, want) {
			t.Fatalf("%s args=%q want %q", runner, plan.agentArgs, want)
		}
		plan.agentArgs = []string{"--open"}
		argv, err := buildAgentArgv(workspace, plan, "task")
		if err != nil {
			t.Fatal(err)
		}
		if !reflect.DeepEqual(argv[4:6], []string{"--open", "--color"}) {
			t.Fatalf("terminal color default changed: %q", argv)
		}
	}
}

func TestDSHWebCreateConfiguredArgs(t *testing.T) {
	for _, tc := range []struct {
		name, args string
		want       []string
	}{
		{"absent", "", []string{"--open", "--no-submit"}},
		{"empty uses defaults", `,"args":[]`, []string{"--open", "--no-submit"}},
		{"custom replaces defaults", `,"args":["--open"]`, []string{"--open"}},
		{"unsupported custom flags are not discarded", `,"args":["--color"]`, []string{"--color"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			config := `{"version":1,"create":{"agent":{"enabled":true,"runner":"dsh-web","prompt_template":"Review ${task}"` + tc.args + `}}}`
			if err := os.WriteFile(filepath.Join(home, "config.json"), []byte(config), 0600); err != nil {
				t.Fatal(err)
			}
			plan, err := resolveCreateUX(home, createUXFlags{}, true)
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(plan.agentArgs, tc.want) {
				t.Fatalf("args=%q want %q", plan.agentArgs, tc.want)
			}
			got, err := buildAgentArgv(home, plan, "task")
			if err != nil {
				t.Fatal(err)
			}
			want := append([]string{"agent-run", "run", "--dir", home}, tc.want...)
			want = append(want, "--agent-runner=dsh-web", "Review task")
			if !reflect.DeepEqual(got, want) {
				t.Fatalf("argv=%q want %q", got, want)
			}
		})
	}
}

func TestDSHWebCreateLongPromptFile(t *testing.T) {
	workspace := t.TempDir()
	// Keep the spill under the test's owned temporary root.
	t.Setenv("TMPDIR", workspace)
	t.Setenv("TMP", workspace)
	t.Setenv("TEMP", workspace)
	task := strings.Repeat("long task 你好\n", 1000)
	got, err := buildAgentArgv(workspace, createUXPlan{
		runner: "dsh-web", promptTmpl: defaultAgentPromptTemplate,
	}, task)
	if err != nil {
		t.Fatal(err)
	}
	wantPrefix := []string{"agent-run", "run", "--dir", workspace, "--open", "--no-submit", "--agent-runner=dsh-web"}
	if len(got) != len(wantPrefix)+1 || !reflect.DeepEqual(got[:len(wantPrefix)], wantPrefix) {
		t.Fatalf("argv=%q", got)
	}
	last := got[len(got)-1]
	if !strings.HasPrefix(last, "--prompt-file=") {
		t.Fatalf("long prompt was not spilled: %q", last)
	}
	path := strings.TrimPrefix(last, "--prompt-file=")
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != strings.TrimSpace("/brainstorm "+task) {
		t.Fatal("spilled task differs from expanded prompt")
	}
}

func TestDSHWebSetConfigDefaults(t *testing.T) {
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, "config.json"), []byte(`{"version":1,"create":{"agent":{"runner":"dsh-web"}}}`), 0600); err != nil {
		t.Fatal(err)
	}
	if err := setConfigWriteCreate(home, createUXFlags{openInAgent: true}); err != nil {
		t.Fatal(err)
	}
	plan, err := resolveCreateUX(home, createUXFlags{}, true)
	if err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(plan.agentArgs, []string{"--open", "--no-submit"}) {
		t.Fatalf("args=%q", plan.agentArgs)
	}
}
