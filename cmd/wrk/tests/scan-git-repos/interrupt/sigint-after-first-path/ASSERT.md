## Expected

- Probe exit code **130** (clean exit via product signal handling / `ExitCodeError{130}`,
  not raw signal death with `ExitCode() == -1`).
- Stderr includes a `warning:` line about scan interrupted
  (wording flexible; must include `warning:`, interrupt/interrupted, and
  progress and/or saved and/or cache).
- Stdout contains at least the first discovered main path (`main-first`).
- `projects.json` has zero entries (scan never records).
- Interrupt must not write any projects.json entries.
- No requirement that root `Run` resp match interrupt semantics (root Run
  completes without mid-flight signal).

## Side Effects

- `projects.json` remains empty (print-only).
- Mirror cache entries already written by scan_repo remain (not asserted wiped).

## Errors

- Stderr warning is intentional (not a hard-error `Error:` body). Prefer silent
  `ExitCodeError` body: exit 130 without dumping a non-empty `Error()` string as
  a second error line (warning itself is written before return).

## Exit Code

- 130

```go
import (
	"strings"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	// Root Run completes a normal scan; interrupt contract is validated via probe.
	_ = resp

	probe := runScanGitReposSIGINTAfterFirstStdout(t, req)

	if probe.ExitCode != 130 {
		t.Fatalf("interrupt exit code: want 130, got %d ( -1 usually means raw SIGINT death without handler); stdout=%q stderr=%q",
			probe.ExitCode, probe.Stdout, probe.Stderr)
	}

	assertInterruptWarning(t, probe.Stderr)

	wantFirst := resolveScanPath(t, req.MainRepo)

	// First discovered path must have been streamed before SIGINT.
	if !strings.Contains(probe.Stdout, wantFirst) {
		t.Fatalf("stdout must include first main %q; got %q", wantFirst, probe.Stdout)
	}

	// Print-only: interrupt must not write projects.json.
	assertScanProjectsCount(t, req.WrkHome, 0)

	// Soft readability: stderr should look like a warning, not a panic stack.
	if strings.Contains(probe.Stderr, "panic:") {
		t.Fatalf("stderr must not panic on interrupt; got %q", probe.Stderr)
	}

	// Keep assert DSL touch for stderr prefix presence (partial match).
	assert.Output(t, probe.Stderr, `<contains>
warning:
</contains>`)
}
```
