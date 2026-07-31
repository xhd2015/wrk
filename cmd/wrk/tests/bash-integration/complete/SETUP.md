# Scenario

**Feature**: wrk --bash-integration --complete returns bash completion candidates

```
seeded projects.json (optional)
wrk --bash-integration --complete -- <words> <cword> -> one candidate per stdout line
```

## Steps

1. Set `req.Mode = "complete"`.
2. Descendants set completion words/cword and optional `req.ProjectPaths`.

```go
import (
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "complete"
	req.DryRun = false
	return nil
}

func seedStandardProjects(req *Request) {
	req.ProjectPaths = []string{
		"/data/alpha",
		"/data/alphalong",
		"/data/beta",
	}
}

func assertCompleteExitOK(t *testing.T, resp *Response) {
	t.Helper()
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0, got %d; stderr=%s stdout=%s", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	assertStdoutEndsWithNewline(t, resp.Stdout)
}

func assertCompleteLines(t *testing.T, stdout string, want []string) {
	t.Helper()
	got := splitCompletionLines(stdout)
	if len(got) != len(want) {
		t.Fatalf("candidate count: got %d %v want %d %v\nstdout=%q", len(got), got, len(want), want, stdout)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("candidate[%d]: got %q want %q\nstdout=%q", i, got[i], want[i], stdout)
		}
	}
}

func splitCompletionLines(stdout string) []string {
	// runBashComplete prints each candidate + a trailing empty line; drop all
	// trailing newlines before splitting so "" is not a spurious candidate.
	trimmed := strings.TrimRight(stdout, "\n")
	if trimmed == "" {
		return nil
	}
	return strings.Split(trimmed, "\n")
}

func assertAllLinesAreFlags(t *testing.T, stdout string) {
	t.Helper()
	for _, line := range splitCompletionLines(stdout) {
		if !strings.HasPrefix(line, "-") {
			t.Fatalf("expected flag candidate, got %q in stdout=%q", line, stdout)
		}
	}
}

func assertFlagsInclude(t *testing.T, stdout string, flags ...string) {
	t.Helper()
	for _, flag := range flags {
		assertContains(t, stdout, flag)
	}
}

// assertExactFlagCandidates requires each flag to appear as a full completion
// line (one candidate per line). Use for short flags like --pr that are
// prefixes of longer flags (e.g. --propagate-tags).
func assertExactFlagCandidates(t *testing.T, stdout string, flags ...string) {
	t.Helper()
	have := map[string]bool{}
	for _, line := range splitCompletionLines(stdout) {
		have[strings.TrimSpace(line)] = true
	}
	for _, flag := range flags {
		if !have[flag] {
			t.Fatalf("completion missing exact candidate %q; stdout=%q", flag, stdout)
		}
	}
}
```
