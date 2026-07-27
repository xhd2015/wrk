---
label: slow
explanation: two scan roots; root-later has 2000 pad dirs + late main; cold stream timing probe
---

## Expected

- Root `Run` (and stream probe): exit code **0**.
- Probe stdout first line is the first discovered main (`main-first` abs path).
- Probe full stdout eventually includes both mains (first then second).
- **Streaming UX**: first stdout bytes arrive at least **40ms** before process exit
  (`total_ms - first_byte_ms >= 40`, with `total_ms >= 80`), proving the first
  path is not held until the entire multi-root scan finishes.
- After probe: `projects.json` does not write projects.json (print-only).

## Side Effects

- Pad dirs are empty and not printed as repos.
- Probe may clear projects.json for isolation; scan still leaves it empty.

## Errors

- No hard errors on the success path; stderr may be empty or non-fatal warnings.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("root Run exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantFirst := resolveScanPath(t, req.MainRepo)
	wantSecond := resolveScanPath(t, req.SecondRepo)

	// Root Run is a full uninterrupted scan (sanity: both paths eventually).
	if !strings.Contains(resp.Stdout, wantFirst) {
		t.Fatalf("root stdout must include first main %q; got %q", wantFirst, resp.Stdout)
	}
	if !strings.Contains(resp.Stdout, wantSecond) {
		t.Fatalf("root stdout must include second main %q; got %q", wantSecond, resp.Stdout)
	}

	probe := runScanGitReposStreamProbe(t, req)
	if probe.ExitCode != 0 {
		t.Fatalf("streaming probe exit %d full=%q", probe.ExitCode, probe.FullStdout)
	}

	assertScanStreamsIncrementally(t, probe, req.MainRepo)

	// Full completion: both mains eventually on stdout (order preserved).
	assert.Output(t, probe.FullStdout, v2StdoutTemplate(wantFirst+"\n"+wantSecond))

	// Print-only: scan never mutates projects.json.
	assertScanProjectsCount(t, req.WrkHome, 0)

	t.Logf("scan stream probe: first_byte_ms=%d total_ms=%d gap_ms=%d",
		probe.FirstByteMS, probe.TotalMS, probe.TotalMS-probe.FirstByteMS)
}
```
