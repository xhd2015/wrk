package wrkcli

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestCreateUXBrowserConfigMergeAndArgv(t *testing.T) {
	home := t.TempDir()
	config := `{"version":1,"create":{"agent":{"enabled":true,"runner":"dsh-web","browser":"brave"}}}`
	if err := os.WriteFile(filepath.Join(home, "config.json"), []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	plan, err := resolveCreateUX(home, createUXFlags{}, true)
	if err != nil {
		t.Fatal(err)
	}
	if plan.browser != "brave" {
		t.Fatalf("browser=%q want brave", plan.browser)
	}
	got, err := buildAgentArgv(home, plan, "task")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"agent-run", "run", "--dir", home, "--open", "--no-submit", "--agent-runner=dsh-web", "--browser=brave", "/brainstorm task"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("argv=%q want %q", got, want)
	}
}

func TestCreateUXBrowserCLIOverridesConfig(t *testing.T) {
	home := t.TempDir()
	config := `{"version":1,"create":{"agent":{"enabled":true,"runner":"dsh-web","browser":"firefox"}}}`
	if err := os.WriteFile(filepath.Join(home, "config.json"), []byte(config), 0600); err != nil {
		t.Fatal(err)
	}
	plan, err := resolveCreateUX(home, createUXFlags{browser: stringPtr("brave")}, true)
	if err != nil {
		t.Fatal(err)
	}
	if plan.browser != "brave" {
		t.Fatalf("browser=%q want brave (CLI must override config)", plan.browser)
	}
}

func TestCreateUXBrowserUnsetOmitsArg(t *testing.T) {
	home := t.TempDir()
	if err := os.WriteFile(filepath.Join(home, "config.json"), []byte(`{"version":1,"create":{"agent":{"enabled":true,"runner":"dsh-web"}}}`), 0600); err != nil {
		t.Fatal(err)
	}
	plan, err := resolveCreateUX(home, createUXFlags{}, true)
	if err != nil {
		t.Fatal(err)
	}
	if plan.browser != "" {
		t.Fatalf("browser=%q want empty", plan.browser)
	}
	got, err := buildAgentArgv(home, plan, "task")
	if err != nil {
		t.Fatal(err)
	}
	for _, a := range got {
		if strings.HasPrefix(a, "--browser") {
			t.Fatalf("argv must not contain --browser without a configured browser: %q", got)
		}
	}
}

func TestCreateUXBrowserRejections(t *testing.T) {
	for _, tc := range []struct {
		name    string
		home    string // empty = no config file
		flags   createUXFlags
		wantErr string
	}{
		{
			name:    "terminal runner via CLI",
			flags:   createUXFlags{openInAgent: true, agentRunner: stringPtr("grok-tty"), browser: stringPtr("brave")},
			wantErr: "wrk: --browser requires the dsh-web runner; current runner is grok-tty",
		},
		{
			name:    "terminal runner via config",
			home:    `{"version":1,"create":{"agent":{"enabled":true,"runner":"grok-tty","browser":"brave"}}}`,
			wantErr: "wrk: --browser requires the dsh-web runner; current runner is grok-tty",
		},
		{
			name:    "agent launch off via CLI",
			flags:   createUXFlags{noOpenInAgent: true, browser: stringPtr("brave")},
			wantErr: "wrk: --browser requires agent launch; remove --no-open-in-agent or pass --open-in-agent",
		},
		{
			name:    "agent launch off via config",
			home:    `{"version":1,"create":{"agent":{"browser":"brave"}}}`,
			wantErr: "wrk: --browser requires agent launch; remove --no-open-in-agent or pass --open-in-agent",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			if tc.home != "" {
				if err := os.WriteFile(filepath.Join(home, "config.json"), []byte(tc.home), 0600); err != nil {
					t.Fatal(err)
				}
			}
			_, err := resolveCreateUX(home, tc.flags, tc.home != "")
			if err == nil || err.Error() != tc.wantErr {
				t.Fatalf("err=%v want %q", err, tc.wantErr)
			}
		})
	}
}

func TestCreateUXBrowserBuildAgentArgvGating(t *testing.T) {
	workspace := t.TempDir()
	for _, tc := range []struct {
		name  string
		plan  createUXPlan
		want  []string
		noBro bool
	}{
		{
			name: "dsh-web appends browser before prompt",
			plan: createUXPlan{runner: "dsh-web", browser: "brave", promptTmpl: defaultAgentPromptTemplate},
			want: []string{"agent-run", "run", "--dir", workspace, "--open", "--no-submit", "--agent-runner=dsh-web", "--browser=brave", "/brainstorm task"},
		},
		{
			name:  "grok-tty never receives browser",
			plan:  createUXPlan{runner: "grok-tty", browser: "brave", promptTmpl: defaultAgentPromptTemplate},
			noBro: true,
			want:  []string{"agent-run", "run", "--dir", workspace, "--session-id-from-prompt", "--no-submit", "--open", "--color", "--agent-runner=grok-tty", "/brainstorm task"},
		},
		{
			name: "dsh-web custom args unchanged with browser",
			plan: createUXPlan{runner: "dsh-web", agentArgs: []string{"--open"}, browser: "brave", promptTmpl: defaultAgentPromptTemplate},
			want: []string{"agent-run", "run", "--dir", workspace, "--open", "--agent-runner=dsh-web", "--browser=brave", "/brainstorm task"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			got, err := buildAgentArgv(workspace, tc.plan, "task")
			if err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("argv=%q want %q", got, tc.want)
			}
			if tc.noBro {
				for _, a := range got {
					if strings.HasPrefix(a, "--browser") {
						t.Fatalf("terminal runner must not receive --browser: %q", got)
					}
				}
			}
		})
	}
}

func TestCreateUXBrowserSetConfigWriteAndClear(t *testing.T) {
	for _, tc := range []struct {
		name      string
		seed      string // empty = no config file
		flags     createUXFlags
		wantBrows string // expected create.agent.browser after write; "absent" = key must be gone
		wantAgent bool   // whether create.agent should still exist
	}{
		{
			name:      "write browser alone",
			flags:     createUXFlags{browser: stringPtr("brave")},
			wantBrows: "brave",
			wantAgent: true,
		},
		{
			name:      "write browser with open-in-agent",
			flags:     createUXFlags{openInAgent: true, browser: stringPtr("brave")},
			wantBrows: "brave",
			wantAgent: true,
		},
		{
			name:      "overwrite existing browser",
			seed:      `{"version":1,"create":{"agent":{"enabled":true,"runner":"dsh-web","browser":"firefox"}}}`,
			flags:     createUXFlags{browser: stringPtr("brave")},
			wantBrows: "brave",
			wantAgent: true,
		},
		{
			name:      "clear browser keeps other agent keys",
			seed:      `{"version":1,"create":{"agent":{"enabled":true,"runner":"dsh-web","browser":"firefox"}}}`,
			flags:     createUXFlags{noBrowser: true},
			wantBrows: "absent",
			wantAgent: true,
		},
		{
			name:      "clear browser removes emptied agent map",
			seed:      `{"version":1,"create":{"agent":{"browser":"firefox"}}}`,
			flags:     createUXFlags{noBrowser: true},
			wantBrows: "absent",
			wantAgent: false,
		},
		{
			name:      "clear browser on empty config is a no-op",
			flags:     createUXFlags{noBrowser: true},
			wantBrows: "absent",
			wantAgent: false,
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			home := t.TempDir()
			if tc.seed != "" {
				if err := os.WriteFile(filepath.Join(home, "config.json"), []byte(tc.seed), 0600); err != nil {
					t.Fatal(err)
				}
			}
			if err := setConfigWriteCreate(home, tc.flags); err != nil {
				t.Fatal(err)
			}
			cfg, err := loadConfig(home)
			if err != nil {
				t.Fatal(err)
			}
			if cfg == nil || cfg.Create == nil {
				t.Fatalf("config.json missing create section: %v", cfg)
			}
			if !tc.wantAgent {
				if cfg.Create.Agent != nil {
					t.Fatalf("create.agent should be removed, got %+v", cfg.Create.Agent)
				}
				return
			}
			if cfg.Create.Agent == nil {
				t.Fatal("create.agent missing after write")
			}
			if tc.wantBrows == "absent" {
				if cfg.Create.Agent.Browser != "" {
					t.Fatalf("browser=%q want cleared", cfg.Create.Agent.Browser)
				}
				data, err := os.ReadFile(filepath.Join(home, "config.json"))
				if err != nil {
					t.Fatal(err)
				}
				if strings.Contains(string(data), "browser") {
					t.Fatalf("config.json still mentions browser: %s", data)
				}
				return
			}
			if cfg.Create.Agent.Browser != tc.wantBrows {
				t.Fatalf("browser=%q want %q", cfg.Create.Agent.Browser, tc.wantBrows)
			}
		})
	}
}

func TestCreateUXBrowserSetConfigParse(t *testing.T) {
	opts, err := parseSetConfigArgs([]string{"--set-config", "--create", "--browser", "brave"})
	if err != nil {
		t.Fatal(err)
	}
	if opts.browser == nil || *opts.browser != "brave" {
		t.Fatalf("browser=%v want brave", opts.browser)
	}
	if !opts.anyCreateFlag() {
		t.Fatal("--browser must count as a create UX flag")
	}
	opts, err = parseSetConfigArgs([]string{"--set-config", "--create", "--no-browser"})
	if err != nil {
		t.Fatal(err)
	}
	if !opts.noBrowser || !opts.anyCreateFlag() {
		t.Fatalf("noBrowser=%v anyCreateFlag=%v", opts.noBrowser, opts.anyCreateFlag())
	}
	for _, args := range [][]string{
		{"--set-config", "--create", "--browser"},
		{"--set-config", "--create", "--browser", ""},
	} {
		if _, err := parseSetConfigArgs(args); err == nil || !strings.Contains(err.Error(), "--browser requires a value") {
			t.Fatalf("args=%q err=%v, want missing-value error", args, err)
		}
	}
}

func TestCreateUXBrowserSetConfigShowAndHelp(t *testing.T) {
	home := t.TempDir()
	if err := setConfigWriteCreate(home, createUXFlags{openInAgent: true, browser: stringPtr("brave")}); err != nil {
		t.Fatal(err)
	}
	res := Capture(CaptureOpts{Args: []string{"--set-config", "--show"}, WrkHome: home})
	if res.ExitCode != 0 {
		t.Fatalf("exit=%d stderr=%q", res.ExitCode, res.Stderr)
	}
	if !strings.Contains(res.Stdout, `"browser": "brave"`) {
		t.Fatalf("--show output missing browser: %s", res.Stdout)
	}
	if !strings.Contains(setConfigCreateUsage(), "--browser NAME") ||
		!strings.Contains(setConfigCreateUsage(), "--no-browser") {
		t.Fatalf("set-config create usage missing browser flags:\n%s", setConfigCreateUsage())
	}
	if !strings.Contains(usage(), "--browser NAME") {
		t.Fatal("wrk usage missing --browser NAME")
	}
}

func TestCreateUXBrowserSelectsCreateOverDashboard(t *testing.T) {
	if isDashboardBareEntry(false, false, nil, createUXFlags{browser: stringPtr("brave")}, false, false, false, false) {
		t.Fatal("--browser must select create over bare dashboard entry")
	}
}

func TestCaptureBrowserSetConfigWriteAndClear(t *testing.T) {
	home := t.TempDir()
	res := Capture(CaptureOpts{Args: []string{"--set-config", "--create", "--open-in-agent", "--browser", "brave"}, WrkHome: home})
	if res.ExitCode != 0 {
		t.Fatalf("write exit=%d stderr=%q", res.ExitCode, res.Stderr)
	}
	cfg, err := loadConfig(home)
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil || cfg.Create == nil || cfg.Create.Agent == nil || cfg.Create.Agent.Browser != "brave" || cfg.Create.Agent.Enabled == nil || !*cfg.Create.Agent.Enabled {
		t.Fatalf("config after write: %+v", cfg)
	}
	res = Capture(CaptureOpts{Args: []string{"--set-config", "--create", "--no-browser"}, WrkHome: home})
	if res.ExitCode != 0 {
		t.Fatalf("clear exit=%d stderr=%q", res.ExitCode, res.Stderr)
	}
	cfg, err = loadConfig(home)
	if err != nil {
		t.Fatal(err)
	}
	if cfg == nil || cfg.Create == nil || cfg.Create.Agent == nil || cfg.Create.Agent.Browser != "" {
		t.Fatalf("config after clear: %+v", cfg)
	}
}

func TestCaptureBrowserRejectsNonCreateMode(t *testing.T) {
	for _, args := range [][]string{
		// --web / --scan-git-repos dispatch before the generic create UX flag
		// guard, so the create-only flag guard must catch them first.
		{"--status", "--browser", "brave"},
		{"--web", "--browser", "brave"},
		{"--scan-git-repos", "--browser", "brave"},
	} {
		res := Capture(CaptureOpts{Args: args})
		if res.ExitCode == 0 || !strings.Contains(res.Stderr, "only valid with create") {
			t.Fatalf("args=%q exit=%d stderr=%q, want create-only error", args, res.ExitCode, res.Stderr)
		}
	}
}
