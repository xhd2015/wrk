## Expected

- Non-zero exit.
- Stderr indicates rejection of combining `--main` with `--gen-commit-msg`.
- Preferred product string: `--main is not valid with --gen-commit-msg` (or equivalent naming both flags).
- Generic `mutually exclusive` is acceptable only if both flags are still named; prefer named-not-valid.

## Errors

- `--main` + `--gen-commit-msg` is illegal.

## Exit Code

- Non-zero

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for --main --gen-commit-msg; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	se := resp.Stderr
	lower := strings.ToLower(se)
	namesMain := strings.Contains(se, "--main")
	namesGen := strings.Contains(se, "--gen-commit-msg") || strings.Contains(se, "gen-commit-msg")
	hasNamedPair := strings.Contains(lower, "not valid with") ||
		strings.Contains(lower, "cannot be used with") ||
		strings.Contains(lower, "not compatible")
	hasMutex := strings.Contains(lower, "mutually exclusive")
	if !(hasNamedPair || hasMutex) {
		t.Fatalf("stderr should reject combination (not valid / mutually exclusive), got %q", se)
	}
	// Target contract: prefer "wrk: --main is not valid with --gen-commit-msg".
	// RED until both flags appear (today may only name one side of a generic mutex).
	if !namesMain || !namesGen {
		t.Fatalf("stderr should name both --main and --gen-commit-msg (named reject preferred), got %q", se)
	}
}
```
