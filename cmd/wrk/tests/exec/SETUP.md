# Scenario

**Feature**: wrk `--exec` runs a trailing command in the mode target directory

```
# less-flags Cut("--exec", &execArgs): tokens after --exec never parsed as wrk flags
# after successful allowed mode → exec.Command(execArgs[0], execArgs[1:]...); cmd.Dir = target abs

wrk --exec pwd
  -> create wt; print wt path; run pwd in wt (stdout path then pwd)

wrk --cd <dir> --exec pwd          # follow-up still written; exec output on stdout
wrk --bring <dep> --exec pwd         # external wt path then pwd there
wrk --set-task <slug> --exec pwd   # rename; print new path then pwd
wrk --done -y --exec pwd           # merge-back --rm; exec in main repo (not removed wt)

# reject / parse
wrk --list --exec true             # non-zero; not valid with this mode
wrk --exec                         # non-zero; requires a command
wrk --exec=pwd                     # non-zero; equals form rejected
```

## Preconditions

- Git (and Go for bring leaves) available; session `wrk` binary via root harness.
- Isolated `WRK_HOME` at `{WorkRoot}/.wrk`; `WRK_DATE=2026-06-30`.
- Leaves put mode flags in `req.Args` (and `SetTaskDesc` / follow-up fields as needed); `--exec` and its command tokens are always last.

## Steps

- Grouping nodes set mode fixtures; leaves append `--exec` (or bare/equals forms) and assert stdout, stderr, exit, and target-dir side effects.
- Success leaves prefer `pwd` (or `echo`) so "ran in dir" is exact-match stdout without flaky contains-only checks.

## Context

- Allowed modes: create (native), `--cd`, `--bring`, `--set-task`, `--done`.
- Rejected with `--exec`: `--list`, `--status`, and other non-allowed modes.
- Mode path lines (create/bring/set-task/done messages) print **before** the child command's stdout.

```go
import (
	"path/filepath"
	"strings"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	ensureExecHelpersUsed()
	return nil
}

// assertPathThenChildStdout expects mode path line(s) then a final exact child line.
// For create/bring/set-task happy path with `pwd`, wantPath appears twice (path print + pwd).
func assertPathThenChildStdout(t *testing.T, stdout, wantPath, childLine string) {
	t.Helper()
	body := wantPath + "\n" + childLine
	assert.Output(t, stdout, v2StdoutTemplate(body))
}

// assertLastStdoutLine checks the last non-empty line of stdout equals want.
func assertLastStdoutLine(t *testing.T, stdout, want string) {
	t.Helper()
	s := strings.TrimSuffix(stdout, "\n")
	if s == "" {
		t.Fatalf("stdout empty; want last line %q", want)
	}
	lines := strings.Split(s, "\n")
	got := lines[len(lines)-1]
	if got != want {
		t.Fatalf("last stdout line: want %q, got %q\nfull stdout:\n%s", want, got, stdout)
	}
}

// resolveAbs canonicalizes a path the same way git/wrk tests do (Abs + EvalSymlinks).
func resolveAbs(t *testing.T, path string) string {
	t.Helper()
	abs, err := filepath.Abs(path)
	if err != nil {
		t.Fatalf("abs %s: %v", path, err)
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return abs
	}
	return resolved
}

func ensureExecHelpersUsed() {
	_ = assertPathThenChildStdout
	_ = assertLastStdoutLine
	_ = resolveAbs
}
```
