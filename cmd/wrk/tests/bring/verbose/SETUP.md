# Scenario

**Feature**: wrk --bring -v logs go mod tidy and streams git worktree add like create

```
# matching dep path (replace + tidy run):
#   -v  -> stderr timestamped $ go -C <modDir> mod tidy + stream tidy child to stderr
#   off -> no tidy pre-line
# new external create under -v:
#   -> stream git worktree add (Preparing worktree / HEAD is now at) after git pre-line
consumer + dep -> wrk --bring <dep> [-v]
```

## Preconditions

- Same fixtures as `bring/basic` (consumer requires dep).
- Verbose logging must not pollute stdout (path line only).

## Steps

- Leaves set `req.Args` with `--bring` and optional `-v`.
- Shared helpers assert tidy pre-line shape and worktree stream markers.

```go
import (
	"regexp"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	_ = req
	ensureBringVerboseHelpersUsed()
	return nil
}

func ensureBringVerboseHelpersUsed() {
	_ = assertBringStderrContainsTidyPreLine
	_ = assertBringStderrNoTidyPreLine
	_ = assertBringStderrContainsWorktreeAddOutput
	_ = assertBringStderrContainsGitWorktreeAdd
}

// assertBringStderrContainsTidyPreLine checks for verbose go mod tidy pre-command log.
// Expected form (flexible on -C placement): timestamp + "$ go" + "mod tidy", preferably with -C and modDir.
func assertBringStderrContainsTidyPreLine(t *testing.T, stderr, modDir string) {
	t.Helper()
	if !strings.Contains(stderr, "mod tidy") {
		t.Fatalf("stderr should contain mod tidy pre-line, got %q", stderr)
	}
	if !strings.Contains(stderr, "$ go") {
		t.Fatalf("stderr should contain $ go tidy pre-line, got %q", stderr)
	}
	re := regexp.MustCompile(`\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\] \$ go `)
	if !re.MatchString(stderr) {
		t.Fatalf("stderr should contain timestamped $ go pre-line, got %q", stderr)
	}
	if modDir != "" {
		if !strings.Contains(stderr, "-C") {
			t.Fatalf("stderr tidy pre-line should include -C, got %q", stderr)
		}
		if !strings.Contains(stderr, modDir) {
			t.Fatalf("stderr tidy pre-line should include module dir %q, got %q", modDir, stderr)
		}
	}
}

// assertBringStderrNoTidyPreLine fails if stderr looks like a verbose go mod tidy pre-line.
func assertBringStderrNoTidyPreLine(t *testing.T, stderr string) {
	t.Helper()
	for _, line := range strings.Split(stderr, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		if strings.Contains(line, "$ go") && strings.Contains(line, "mod tidy") {
			t.Fatalf("stderr should not contain tidy pre-line without -v, got %q", stderr)
		}
		if strings.Contains(line, "mod tidy") && strings.Contains(line, "$ go") {
			t.Fatalf("stderr should not contain tidy pre-line without -v, got %q", stderr)
		}
	}
	// Also reject bare "mod tidy" on a verbose pre-line style line.
	if strings.Contains(stderr, "mod tidy") && strings.Contains(stderr, "$ go") {
		t.Fatalf("stderr should not contain $ go … mod tidy without -v, got %q", stderr)
	}
}

// assertBringStderrContainsWorktreeAddOutput mirrors fetch-and-verbose create stream asserts.
func assertBringStderrContainsWorktreeAddOutput(t *testing.T, stderr string) {
	t.Helper()
	if strings.Contains(stderr, "Preparing worktree") || strings.Contains(stderr, "HEAD is now at") {
		return
	}
	t.Fatalf("stderr should contain git worktree add subprocess output (Preparing worktree or HEAD is now at), got %q", stderr)
}

// assertBringStderrContainsGitWorktreeAdd checks for verbose git worktree add pre-line.
func assertBringStderrContainsGitWorktreeAdd(t *testing.T, stderr string) {
	t.Helper()
	if !strings.Contains(stderr, "git ") || !strings.Contains(stderr, "worktree add") {
		t.Fatalf("stderr should contain git worktree add log line, got %q", stderr)
	}
	re := regexp.MustCompile(`\[\d{4}-\d{2}-\d{2} \d{2}:\d{2}:\d{2}\] \$ git `)
	if !re.MatchString(stderr) {
		t.Fatalf("stderr should contain timestamped git log line, got %q", stderr)
	}
}
```
