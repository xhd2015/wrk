## Expected

- Exit code 0.
- Flag layer accepts `--gen-commit-msg` + `--sync` without done/merge-back.
- Stderr must **not** contain `mutually exclusive`.
- Gen-commit dry plan visible (`would generate` / dry-run vocabulary).
- Sync stage runs (empty distribute is OK; prefer `would: synced:` or sync summary).

## Side Effects

- None under `--dry-run` (staged file remains staged).

## Exit Code

- 0

```go
import (
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	se := resp.Stderr
	if strings.Contains(se, "mutually exclusive") {
		t.Fatalf("bare --gen-commit-msg --sync still mutually exclusive; stderr=%q exit=%d",
			se, resp.ExitCode)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 for gen-commit+sync dry-run compose; exit=%d stdout=%q stderr=%q",
			resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	outAll := resp.Stdout + "\n" + resp.Stderr
	if !strings.Contains(outAll, "would generate") &&
		!strings.Contains(strings.ToLower(outAll), "dry-run") {
		t.Fatalf("expected gen-commit dry-run plan; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	// Sync stage should at least plan a summary (0/0/0 is fine on a lone main).
	if !strings.Contains(outAll, "would: synced:") &&
		!strings.Contains(outAll, "synced:") {
		t.Fatalf("expected sync stage plan/summary; stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
}
```
