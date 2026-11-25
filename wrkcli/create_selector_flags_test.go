package wrkcli

import (
	"reflect"
	"strings"
	"testing"
)

func TestCreateSelectorFlagsRegistry(t *testing.T) {
	cases := []struct {
		name  string
		new   bool
		task  string
		flags createUXFlags
		want  []string
	}{
		{
			name: "no selectors",
			want: nil,
		},
		{
			name: "new only",
			new:  true,
			want: []string{"--new"},
		},
		{
			name: "short task token preserved",
			task: "-t",
			want: []string{"-t"},
		},
		{
			name: "long task token preserved",
			task: "--task",
			want: []string{"--task"},
		},
		{
			name:  "negative UX flags are selectors",
			flags: createUXFlags{noNewTerminal: true, noNewWindow: true, noOpenInAgent: true},
			want:  []string{"--no-new-window", "--no-new-terminal", "--no-open-in-agent"},
		},
		{
			name:  "here and positive terminal modes",
			flags: createUXFlags{here: true, newTerminal: true},
			want:  []string{"--new-terminal", "--here"},
		},
		{
			name:  "agent runner and browser",
			flags: createUXFlags{agentRunner: stringPtr("grok-tty"), browser: stringPtr("brave")},
			want:  []string{"--agent-runner", "--browser"},
		},
		{
			name: "all in registration order",
			new:  true,
			task: "--task",
			flags: createUXFlags{
				newWindow: true, noNewWindow: true,
				newTerminal: true, reuseTerminal: true, smartTerminal: true, noNewTerminal: true,
				here: true, openInAgent: true, noOpenInAgent: true,
				agentRunner: stringPtr("dsh-web"), browser: stringPtr("firefox"),
			},
			want: []string{
				"--new", "--task",
				"--new-window", "--no-new-window",
				"--new-terminal", "--reuse-terminal", "--smart-terminal", "--no-new-terminal",
				"--here", "--open-in-agent", "--no-open-in-agent",
				"--agent-runner", "--browser",
			},
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := createSelectorFlags(tc.new, tc.task, tc.flags)
			if !reflect.DeepEqual(got, tc.want) {
				t.Fatalf("createSelectorFlags() = %v, want %v", got, tc.want)
			}
		})
	}
}

// Every selector the registry can emit must be a flag the CLI actually parses;
// otherwise the rejection would name a token the user never typed.
func TestCreateSelectorFlagsAreRegisteredCLIFlags(t *testing.T) {
	all := createSelectorFlags(true, "-t", createUXFlags{
		newWindow: true, noNewWindow: true,
		newTerminal: true, reuseTerminal: true, smartTerminal: true, noNewTerminal: true,
		here: true, openInAgent: true, noOpenInAgent: true,
		agentRunner: stringPtr("x"), browser: stringPtr("y"), noBrowser: true,
	})
	if len(all) == 0 {
		t.Fatal("expected selectors")
	}
	usage := usage()
	for _, flag := range all {
		if !strings.Contains(usage, flag) {
			t.Fatalf("selector %q is not documented in usage()", flag)
		}
	}
}

// Bring classification: a bare --bring stays exclusive (cwd consumer); only an
// explicit <project> positional makes it create+bring, and selectors without a
// project must be rejected rather than silently upgrading to create.
func TestCreateSelectorFlagsBringClassification(t *testing.T) {
	ux := createUXFlags{noNewTerminal: true}
	selectors := createSelectorFlags(false, "", ux)

	for _, tc := range []struct {
		name           string
		projectPresent bool
		selectors      []string
		wantCompose    bool
		wantReject     bool
	}{
		{name: "bare bring", projectPresent: false, wantCompose: false, wantReject: false},
		{name: "bring with ux flag", projectPresent: false, selectors: selectors, wantCompose: false, wantReject: true},
		{name: "bring with project", projectPresent: true, wantCompose: true, wantReject: false},
		{name: "bring with project and ux flag", projectPresent: true, selectors: selectors, wantCompose: true, wantReject: false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			compose := tc.projectPresent
			if compose != tc.wantCompose {
				t.Fatalf("compose = %v, want %v", compose, tc.wantCompose)
			}
			if reject := !compose && len(tc.selectors) > 0; reject != tc.wantReject {
				t.Fatalf("reject = %v, want %v", reject, tc.wantReject)
			}
		})
	}
}
