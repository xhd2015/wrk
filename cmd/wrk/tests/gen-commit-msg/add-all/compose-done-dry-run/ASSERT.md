## Expected

- **Must not** report lessflags/wrk `unrecognized flag: --add-all` (or `unknown flag` for `--add-all`).
  That is the Classic RED pin until `genCommitMsgBoolFlags` peels `--add-all`.
- Exit code **0** after peel is implemented (full dry-run plan).
- Stderr contains dry-run plan line `would: git add -A`.
- Stderr does **not** log real `$ git add -A`.
- Gen-commit dry plan visible: mock B (`would generate commit message`) and/or `would:` + `git commit`.
- Stderr must **not** contain `mutually exclusive`.
- Worktree HEAD subject unchanged (no real commit under dry-run).

## Side Effects

- Plan only: no real `git add -A`, no commit, no merge-back apply.
- Staged file may remain staged after dry-run.

## Exit Code

- 0 (after peel GREEN)

```go
import (
	"strings"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)

	combined := resp.Stdout + "\n" + resp.Stderr
	lower := strings.ToLower(combined)

	// Classic RED pin: peel must remove --add-all before lessflags.
	if strings.Contains(lower, "unrecognized flag") || strings.Contains(lower, "unknown flag") {
		if strings.Contains(combined, "--add-all") {
			t.Fatalf("compose must peel --add-all (not leave it for lessflags); exit=%d stdout=%q stderr=%q",
				resp.ExitCode, resp.Stdout, resp.Stderr)
		}
	}
	if strings.Contains(resp.Stderr, "mutually exclusive") {
		t.Fatalf("compose --gen-commit-msg --add-all --commit --done must not be mutually exclusive; stderr=%q",
			resp.Stderr)
	}

	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 for --gen-commit-msg --add-all --commit --done --dry-run, got %d stdout=%q stderr=%q",
			resp.ExitCode, resp.Stdout, resp.Stderr)
	}

	if !strings.Contains(resp.Stderr, "would: git add -A") {
		t.Fatalf("stderr must contain would: git add -A after peel+forward, got stderr:\n%s\nstdout:\n%s",
			resp.Stderr, resp.Stdout)
	}
	if strings.Contains(resp.Stderr, "$ git add -A") {
		t.Fatalf("dry-run must not log real $ git add -A, stderr:\n%s", resp.Stderr)
	}

	// Gen-commit dry plan (mock B and/or would: git commit).
	hasMockB := strings.Contains(combined, "would generate commit message")
	hasWouldCommit := strings.Contains(strings.ToLower(resp.Stderr), "would:") &&
		strings.Contains(resp.Stderr, "git commit")
	if !hasMockB && !hasWouldCommit {
		t.Fatalf("expected gen-commit dry-run plan (mock B and/or would: git commit); stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	if strings.Contains(resp.Stderr, "Running git commit...") {
		t.Fatalf("must not execute git commit under dry-run, stderr:\n%s", resp.Stderr)
	}

	if req.HEADSubject != "" {
		got := gitHEADSubject(t, req.RepoDir)
		if got != req.HEADSubject {
			t.Fatalf("worktree HEAD subject changed under dry-run: before=%q after=%q", req.HEADSubject, got)
		}
	}

	assert.Output(t, resp.Stderr, `<contains>
would: git add -A
</contains>`)
}
```
