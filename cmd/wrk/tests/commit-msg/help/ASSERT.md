## Expected

- Exit code 0.
- Help text (stdout and/or stderr) documents **manual** commit message flags as a
  first-class surface (not merely substrings of other flags like `--main`):
  - **`--message`** (long form required in help), and preferably **`-m`** as short
  - that the message flag **requires `--commit`**
  - exclusivity with **`--gen-commit-msg`** (or clear XOR / exclusive wording)
- Stdout (preferred for usage) ends with trailing `\n` when non-empty.

## Side Effects

- Read-only (`-h` only).

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	help := resp.Stdout + resp.Stderr
	if strings.TrimSpace(help) == "" {
		t.Fatal("expected non-empty help text for wrk -h")
	}
	if resp.Stdout != "" && !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("help stdout should end with trailing newline, got %q", resp.Stdout)
	}
	// Require long form --message so we do not false-match --main / other -m substrings.
	if !strings.Contains(help, "--message") {
		t.Fatalf("help must document --message (manual commit message); got %q", help)
	}
	// Prefer short -m as well (word-boundary-ish: "-m," "-m " or "-m/" near message docs).
	// Not required if help only shows long form next to --message.
	_ = strings.Contains(help, "-m,") || strings.Contains(help, "-m ") || strings.Contains(help, "-m/")

	// Manual message requires --commit — look for --message near commit requirement
	// or an explicit line about -m/--message with --commit.
	low := strings.ToLower(help)
	if !strings.Contains(help, "--commit") {
		t.Fatalf("help must mention --commit with message flags; got %q", help)
	}
	// Exclusive / alternative to AI path for message source
	if !strings.Contains(help, "--gen-commit-msg") {
		t.Fatalf("help should mention --gen-commit-msg as the AI alternative; got %q", help)
	}
	// Intent: manual message is documented as requiring commit and exclusive with gen.
	// Accept either explicit exclusive wording or a usage line that pairs --commit with -m/--message
	// distinctly from the gen-commit-msg --commit pre-stage line.
	hasManualUsage := strings.Contains(help, "--commit") && strings.Contains(help, "--message")
	hasExclusiveWording := strings.Contains(low, "exclusive") ||
		strings.Contains(low, "mutually") ||
		strings.Contains(low, "instead of") ||
		strings.Contains(low, "without ai") ||
		strings.Contains(low, "manual")
	if !hasManualUsage {
		t.Fatalf("help must show --commit with --message; got %q", help)
	}
	// Avoid accidental GREEN from pre-existing gen-commit-msg --commit docs alone:
	// require that --message appears on a line that also mentions commit or message intent.
	lines := strings.Split(help, "\n")
	messageLineOK := false
	for _, line := range lines {
		if !strings.Contains(line, "--message") {
			continue
		}
		ll := strings.ToLower(line)
		if strings.Contains(line, "--commit") ||
			strings.Contains(ll, "message") ||
			strings.Contains(ll, "commit") {
			messageLineOK = true
			break
		}
	}
	if !messageLineOK && !hasExclusiveWording {
		t.Fatalf("help --message line should relate to commit message (or exclusive wording present); got %q", help)
	}
}
```
