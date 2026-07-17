## Expected

- Probe exit code **130** (clean exit via product signal handling / `ExitCodeError{130}`,
  not raw signal death with `ExitCode() == -1`).
- Stderr includes a `warning:` line about scan interrupted and progress saved
  (wording flexible; must include `warning:`, interrupt/interrupted, and
  progress and/or saved).
- Stdout contains at least the first discovered main path (`main-first`).
- `projects.json` records `main-first` with `source: "scan"` (progress kept).
- Second main (`zzz-main`) is absent from `projects.json` when the harness
  interrupted before it was printed; if it appears on probe stdout it may also
  be recorded (consistency). Prefer absent under the pad layout.
- No requirement that root `Run` resp match interrupt semantics (root Run
  completes without mid-flight signal).

## Side Effects

- Partial `projects.json` retained (already-recorded mains kept).
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
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
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
	wantSecond := resolveScanPath(t, req.SecondRepo)

	// First newly recorded path must have been streamed before SIGINT.
	if !strings.Contains(probe.Stdout, wantFirst) {
		t.Fatalf("stdout must include first main %q; got %q", wantFirst, probe.Stdout)
	}

	// Progress saved: first main stays in projects.json with source=scan.
	assertScanProjectRecorded(t, req.WrkHome, req.MainRepo, "scan")

	// Partial: if second was never printed, it must not be recorded yet.
	if !strings.Contains(probe.Stdout, wantSecond) {
		assertScanProjectNotRecorded(t, req.WrkHome, req.SecondRepo)
	}

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
