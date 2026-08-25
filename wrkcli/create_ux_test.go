package wrkcli

import (
	"strings"
	"testing"
)

func stringPtr(s string) *string { return &s }

func TestResolveCreateUXAgentRunnerOverride(t *testing.T) {
	plan, err := resolveCreateUX(t.TempDir(), createUXFlags{
		openInAgent: true,
		agentRunner: stringPtr("codex"),
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.agent {
		t.Fatal("agent should be enabled by --open-in-agent")
	}
	if plan.runner != "codex-tty" {
		t.Fatalf("runner=%q want codex-tty", plan.runner)
	}
}

func TestResolveCreateUXAgentRunnerAcceptedValues(t *testing.T) {
	for _, tc := range []struct{ in, want string }{
		{"codex", "codex-tty"},
		{"codex-tty", "codex-tty"},
		{"grok", "grok-tty"},
		{"grok-tty", "grok-tty"},
	} {
		t.Run(tc.in, func(t *testing.T) {
			got, err := normalizeCreateAgentRunner(tc.in)
			if err != nil || got != tc.want {
				t.Fatalf("normalizeCreateAgentRunner(%q) = %q, %v; want %q, nil", tc.in, got, err, tc.want)
			}
		})
	}
}

func TestResolveCreateUXAgentRunnerRequiresAgent(t *testing.T) {
	_, err := resolveCreateUX(t.TempDir(), createUXFlags{agentRunner: stringPtr("codex")}, false)
	if err == nil || !strings.Contains(err.Error(), "requires agent launch") {
		t.Fatalf("err=%v, want agent launch requirement", err)
	}
}

func TestResolveCreateUXAgentRunnerRejectsUnsupportedValue(t *testing.T) {
	_, err := resolveCreateUX(t.TempDir(), createUXFlags{
		openInAgent: true,
		agentRunner: stringPtr("opencode"),
	}, false)
	if err == nil || !strings.Contains(err.Error(), "unsupported create agent runner") {
		t.Fatalf("err=%v, want unsupported runner error", err)
	}
}

func TestResolveCreateUXCodexRunnerUsesBrainstormSkill(t *testing.T) {
	// CLI flag: --agent-runner codex (normalized to codex-tty)
	plan, err := resolveCreateUX(t.TempDir(), createUXFlags{
		openInAgent: true,
		agentRunner: stringPtr("codex"),
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	if plan.promptTmpl != codexAgentPromptTemplate {
		t.Fatalf("promptTmpl=%q want %q", plan.promptTmpl, codexAgentPromptTemplate)
	}
}

func TestResolveCreateUXCodexTtyRunnerUsesBrainstormSkill(t *testing.T) {
	plan, err := resolveCreateUX(t.TempDir(), createUXFlags{
		openInAgent: true,
		agentRunner: stringPtr("codex-tty"),
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	if plan.promptTmpl != codexAgentPromptTemplate {
		t.Fatalf("promptTmpl=%q want %q", plan.promptTmpl, codexAgentPromptTemplate)
	}
}

func TestResolveCreateUXGrokRunnerKeepsSlashBrainstorm(t *testing.T) {
	plan, err := resolveCreateUX(t.TempDir(), createUXFlags{
		openInAgent: true,
		agentRunner: stringPtr("grok"),
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	if plan.promptTmpl != defaultAgentPromptTemplate {
		t.Fatalf("promptTmpl=%q want %q", plan.promptTmpl, defaultAgentPromptTemplate)
	}
}

func TestResolveCreateUXDefaultRunnerUsesSlashBrainstorm(t *testing.T) {
	plan, err := resolveCreateUX(t.TempDir(), createUXFlags{
		openInAgent: true,
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	if plan.promptTmpl != defaultAgentPromptTemplate {
		t.Fatalf("promptTmpl=%q want %q", plan.promptTmpl, defaultAgentPromptTemplate)
	}
}

func TestCaptureAgentRunnerRejectsNonCreateMode(t *testing.T) {
	res := Capture(CaptureOpts{Args: []string{"--status", "--agent-runner", "codex"}})
	if res.ExitCode == 0 || !strings.Contains(res.Stderr, "only valid with create") {
		t.Fatalf("exit=%d stderr=%q, want create-only error", res.ExitCode, res.Stderr)
	}
}

func TestResolveCreateUXHereClearsWindowAndTerminal(t *testing.T) {
	plan, err := resolveCreateUX(t.TempDir(), createUXFlags{
		here:        true,
		openInAgent: true,
	}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.here {
		t.Fatal("here should be set")
	}
	if plan.window {
		t.Fatal("here should clear window")
	}
	if plan.terminalMode != "" {
		t.Fatalf("terminalMode=%q want empty", plan.terminalMode)
	}
	if !plan.agent {
		t.Fatal("open-in-agent should still enable agent")
	}
}

func TestResolveCreateUXHereDoesNotImplyAgent(t *testing.T) {
	plan, err := resolveCreateUX(t.TempDir(), createUXFlags{here: true}, false)
	if err != nil {
		t.Fatal(err)
	}
	if !plan.here {
		t.Fatal("here should be set")
	}
	if plan.agent {
		t.Fatal("--here alone must not enable agent")
	}
}

func TestCreateUXFlagsHereConflictsWithTerminal(t *testing.T) {
	err := createUXFlags{here: true, newTerminal: true}.validate()
	if err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("err=%v, want terminal/no-new-terminal mutual exclusion", err)
	}
}

func TestCreateUXFlagsHereConflictsWithNewWindow(t *testing.T) {
	err := createUXFlags{here: true, newWindow: true}.validate()
	if err == nil || !strings.Contains(err.Error(), "mutually exclusive") {
		t.Fatalf("err=%v, want window mutual exclusion", err)
	}
}

func TestCreateUXFlagsHereAllowsRedundantNoNewFlags(t *testing.T) {
	err := createUXFlags{here: true, noNewWindow: true, noNewTerminal: true, openInAgent: true}.validate()
	if err != nil {
		t.Fatalf("redundant --no-new-* with --here should be ok: %v", err)
	}
}
