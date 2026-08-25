# Scenario

**Feature**: nested `-h`/`--help` for `wrk --bash-integration` (usage, not script dump)

```
# help short-circuits before print / install / uninstall / status
wrk --bash-integration -h|--help
  -> dedicated bash-integration usage on stdout
  -> exit 0, empty stderr, trailing newline
  -> stdout is NOT the bash.sh script

wrk --bash-integration --install -h|--help
  -> same usage; no profile/script writes

# help may appear before the action flag
wrk --bash-integration --help --install -> usage, no write

# root help points at bash-integration
wrk --help -> mentions Bash integration / --bash-integration --help
```

## Preconditions

- Isolated `{WRK_HOME}` + fake `{HOME}` per leaf (root bash-integration `Setup`).
- Help must not print the integration script or mutate profiles/script.
- Prefer `req.InProcess = true` (L2).

## Steps

- Grouping nodes choose help shape (`alone` / `with-install` / `root-help`).
- Leaves set `req.CLIArgs` for the help form under test.
- Asserts pin process contract and usage tokens (not sealed golden strings).

## Context

- Split factor: **help surface** (dedicated mode help vs root pointer vs help+mutating action).
- Within dedicated help, siblings split on **help form** (`--help` vs `-h`) and action co-presence / order.
- `--complete` stays hidden from usage (not asserted as present).

```go
import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.RepoDir == "" {
		req.RepoDir = req.WorkRoot
	}
	ensureBashIntegrationHelpHelpersUsed()
	return nil
}

func assertBashIntegrationHelpProcessContract(t *testing.T, stdout, stderr string, exitCode int) {
	t.Helper()
	if exitCode != 0 {
		t.Fatalf("expected exit 0 for bash-integration help, got %d stderr=%q stdout=%q", exitCode, stderr, stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("stderr should be empty for bash-integration help, got %q", stderr)
	}
	if strings.TrimSpace(stdout) == "" {
		t.Fatal("expected non-empty bash-integration help on stdout")
	}
	if !strings.HasSuffix(stdout, "\n") {
		t.Fatalf("bash-integration help stdout must end with trailing newline, got %q", stdout)
	}
}

func assertBashIntegrationUsage(t *testing.T, stdout, stderr string, exitCode int) {
	t.Helper()
	assertBashIntegrationHelpProcessContract(t, stdout, stderr, exitCode)
	lower := strings.ToLower(stdout)
	for _, want := range []string{"--bash-integration", "--install", "--uninstall", "--status"} {
		if !strings.Contains(lower, want) {
			t.Fatalf("bash-integration help missing %q:\n%s", want, stdout)
		}
	}
	if !strings.Contains(lower, "--help") && !strings.Contains(lower, "-h") {
		t.Fatalf("bash-integration help must mention -h or --help:\n%s", stdout)
	}
	// Must not dump the integration script.
	if strings.HasPrefix(strings.TrimSpace(stdout), "#!/usr/bin/env bash") {
		t.Fatalf("bash-integration help must not dump bash.sh script:\n%s", stdout)
	}
	if strings.Contains(stdout, "complete -o default -F _wrk wrk") {
		t.Fatalf("bash-integration help must not dump bash.sh complete line:\n%s", stdout)
	}
	// Hidden callback stays out of usage.
	if strings.Contains(lower, "--complete") {
		t.Fatalf("bash-integration help must not document hidden --complete:\n%s", stdout)
	}
}

func assertBashIntegrationHelpNoFilesystemWrite(t *testing.T, req *Request) {
	t.Helper()
	script := bashShPath(req.WrkHome)
	if _, err := os.Stat(script); !os.IsNotExist(err) {
		t.Fatalf("help must not write bash.sh at %s (err=%v)", script, err)
	}
	for _, name := range []string{".bash_profile", ".bashrc"} {
		p := filepath.Join(req.FakeHome, name)
		data, ok := readFileIfExists(p)
		if ok && strings.Contains(data, wrkMarkerBegin) {
			t.Fatalf("help must not write profile marker in %s", p)
		}
	}
}

func ensureBashIntegrationHelpHelpersUsed() {
	_ = assertBashIntegrationHelpProcessContract
	_ = assertBashIntegrationUsage
	_ = assertBashIntegrationHelpNoFilesystemWrite
}
```
