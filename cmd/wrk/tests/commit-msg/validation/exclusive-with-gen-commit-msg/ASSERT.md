## Expected

- Non-zero exit code.
- Stderr indicates mutual exclusion / exclusivity between `--gen-commit-msg` and `-m`/`--message`.
- Prefer mentioning both sides of the XOR when product wording allows.
- Must **not** be a late-path git error (`not a git repository`) or bare `unrecognized flag`
  without exclusive intent — those are accidental GREEN before the feature lands.

## Side Effects

- No agent invocation; no commit.

## Errors

- `--gen-commit-msg` cannot be combined with `-m`/`--message`.

## Exit Code

- Non-zero

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitNonZero(t, resp)
	errText := resp.Stderr
	low := strings.ToLower(errText)
	// Reject accidental GREEN before XOR validation exists.
	if strings.Contains(errText, "unrecognized flag") {
		t.Fatalf("expected exclusive reject for gen + --message, got unrecognized flag: %q", errText)
	}
	if strings.Contains(low, "not a git repository") {
		t.Fatalf("expected early exclusive reject for gen + --message, got late git error: %q", errText)
	}
	exclusive := strings.Contains(low, "mutually exclusive") ||
		strings.Contains(low, "exclusive") ||
		strings.Contains(low, "cannot be used with") ||
		strings.Contains(low, "not valid with") ||
		strings.Contains(low, "incompatible")
	if !exclusive {
		t.Fatalf("stderr should indicate exclusive/mutex between gen and -m/--message, got %q", errText)
	}
	if !strings.Contains(errText, "--gen-commit-msg") && !strings.Contains(low, "gen-commit") {
		t.Fatalf("stderr should mention --gen-commit-msg, got %q", errText)
	}
	if !strings.Contains(errText, "-m") && !strings.Contains(errText, "--message") {
		t.Fatalf("stderr should mention -m or --message, got %q", errText)
	}
}
```
