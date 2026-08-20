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

func TestCaptureAgentRunnerRejectsNonCreateMode(t *testing.T) {
	res := Capture(CaptureOpts{Args: []string{"--status", "--agent-runner", "codex"}})
	if res.ExitCode == 0 || !strings.Contains(res.Stderr, "only valid with create") {
		t.Fatalf("exit=%d stderr=%q, want create-only error", res.ExitCode, res.Stderr)
	}
}
