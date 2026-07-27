# Scenario

**Feature**: nested level-specific `-h`/`--help` for `wrk --set-config` (dispatcher, create, show)

```
# help short-circuits before write / UX-flag requirements
wrk --set-config -h|--help
  -> set-config dispatcher usage (actions --create / --show)
  -> exit 0, empty stderr, trailing newline

wrk --set-config --create -h|--help
  -> dedicated create usage (UX flags, conflicts, examples)
  -> exit 0; no config.json write

wrk --set-config --show -h|--help
  -> show usage (pretty-print config.json)
  -> exit 0; help body is not dumped config JSON

# help may appear in any order among tokens
wrk --set-config --help --create -> create-level help
wrk --set-config --create --new-window --help -> create help, still no write
```

## Preconditions

- Isolated `{WRK_HOME}` per leaf (root `Setup`); management does not need git.
- Help must not create or merge `config.json` solely because help was requested.
- Current production parser treats `--help`/`-h` as unrecognized — leaves are RED until implementer lands.

## Steps

- Grouping nodes choose help level (`set-config` / `create` / `show`).
- Leaves set `req.Args` via `setConfigArgs(...)` for the help form under test.
- Asserts pin process contract (exit 0, empty stderr, trailing `\n`) and level-specific tokens.

## Context

- Split factor at this node: **help level** (most significant for nested help).
- Within each level, siblings split on **help form** (`--help` vs `-h`) and optional order / UX co-presence.
- Token asserts (not sealed golden strings) mirror `skill/help` + `assertSkillUsageStdout`.
- Create help must be **dedicated** (rich UX detail), not a copy of the dispatcher page.
- Dispatcher help must point users toward action-level help and list both actions.

```go
import (
	"os"
	"strings"
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	if req.RepoDir == "" {
		req.RepoDir = req.WorkRoot
	}
	ensureSetConfigHelpHelpersUsed()
	return nil
}

func assertSetConfigHelpProcessContract(t *testing.T, stdout, stderr string, exitCode int) {
	t.Helper()
	if exitCode != 0 {
		t.Fatalf("expected exit 0 for set-config help, got %d stderr=%q stdout=%q", exitCode, stderr, stdout)
	}
	if strings.TrimSpace(stderr) != "" {
		t.Fatalf("stderr should be empty for set-config help, got %q", stderr)
	}
	if strings.TrimSpace(stdout) == "" {
		t.Fatal("expected non-empty set-config help on stdout")
	}
	if !strings.HasSuffix(stdout, "\n") {
		t.Fatalf("set-config help stdout must end with trailing newline, got %q", stdout)
	}
}

func assertSetConfigDispatcherHelp(t *testing.T, stdout, stderr string, exitCode int) {
	t.Helper()
	assertSetConfigHelpProcessContract(t, stdout, stderr, exitCode)
	lower := strings.ToLower(stdout)
	for _, want := range []string{"--set-config", "--create", "--show"} {
		if !strings.Contains(lower, want) {
			t.Fatalf("set-config dispatcher help missing %q:\n%s", want, stdout)
		}
	}
	if !strings.Contains(lower, "--help") && !strings.Contains(lower, "-h") {
		t.Fatalf("set-config dispatcher help must mention -h or --help:\n%s", stdout)
	}
	// Pointer toward create-level help (exact wording implementation-owned).
	if !strings.Contains(lower, "--create --help") &&
		!strings.Contains(lower, "--create -h") &&
		!strings.Contains(lower, "create --help") &&
		!strings.Contains(lower, "create -h") {
		t.Fatalf("set-config dispatcher help should point users to create-level help (e.g. --create --help):\n%s", stdout)
	}
	// Must not be the dedicated create page: omit several create UX flags.
	createOnlyUX := []string{"--new-window", "--open-in-agent", "--no-open-in-agent", "--new-terminal"}
	uxHits := 0
	for _, u := range createOnlyUX {
		if strings.Contains(lower, u) {
			uxHits++
		}
	}
	if uxHits >= 3 {
		t.Fatalf("set-config dispatcher help looks like create-level UX dump (found %d UX flags); keep dispatcher brief:\n%s", uxHits, stdout)
	}
}

func assertSetConfigCreateHelp(t *testing.T, stdout, stderr string, exitCode int) {
	t.Helper()
	assertSetConfigHelpProcessContract(t, stdout, stderr, exitCode)
	lower := strings.ToLower(stdout)
	for _, want := range []string{"--set-config", "--create", "--new-window", "--open-in-agent"} {
		if !strings.Contains(lower, want) {
			t.Fatalf("set-config create help missing %q:\n%s", want, stdout)
		}
	}
	// Dedicated create page: more UX detail than dispatcher (negatives and/or terminal modes).
	extraUX := []string{
		"--no-open-in-agent",
		"--no-new-window",
		"--new-terminal",
		"--reuse-terminal",
		"--smart-terminal",
		"--no-new-terminal",
	}
	extraHits := 0
	for _, u := range extraUX {
		if strings.Contains(lower, u) {
			extraHits++
		}
	}
	if extraHits < 2 {
		t.Fatalf("set-config create help should list multiple UX flags beyond --new-window/--open-in-agent (found %d of %v):\n%s", extraHits, extraUX, stdout)
	}
	// Distinct from show help: create page is not a show-only blurb.
	if !strings.Contains(lower, "--create") {
		t.Fatalf("create help must mention --create:\n%s", stdout)
	}
}

func assertSetConfigShowHelp(t *testing.T, stdout, stderr string, exitCode int) {
	t.Helper()
	assertSetConfigHelpProcessContract(t, stdout, stderr, exitCode)
	lower := strings.ToLower(stdout)
	for _, want := range []string{"--set-config", "--show"} {
		if !strings.Contains(lower, want) {
			t.Fatalf("set-config show help missing %q:\n%s", want, stdout)
		}
	}
	// JSON / config / pretty-print theme (token flexible).
	themeOK := strings.Contains(lower, "json") ||
		strings.Contains(lower, "config") ||
		strings.Contains(lower, "pretty") ||
		strings.Contains(lower, "{}")
	if !themeOK {
		t.Fatalf("set-config show help should mention JSON/config/pretty-print theme:\n%s", stdout)
	}
	// Must not dump actual config content as the help body (pure JSON object).
	trimmed := strings.TrimSpace(stdout)
	if strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}") &&
		!strings.Contains(lower, "usage") && !strings.Contains(lower, "help") {
		t.Fatalf("set-config show help looks like dumped config JSON, not usage:\n%s", stdout)
	}
	if !strings.Contains(lower, "--help") && !strings.Contains(lower, "-h") && !strings.Contains(lower, "usage") {
		t.Fatalf("set-config show help should read as usage (mention help/usage):\n%s", stdout)
	}
}

func assertSetConfigNoConfigWrite(t *testing.T, wrkHome string) {
	t.Helper()
	path := setConfigPath(wrkHome)
	if _, err := os.Stat(path); err == nil {
		t.Fatalf("help must not write config.json; found %s", path)
	} else if !os.IsNotExist(err) {
		t.Fatalf("stat %s: %v", path, err)
	}
}

func ensureSetConfigHelpHelpersUsed() {
	_ = assertSetConfigHelpProcessContract
	_ = assertSetConfigDispatcherHelp
	_ = assertSetConfigCreateHelp
	_ = assertSetConfigShowHelp
	_ = assertSetConfigNoConfigWrite
}
```
