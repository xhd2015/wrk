## Expected

- Exit code 0.
- Help text (stdout and/or stderr) documents flag tokens:
  - `--pr`
  - `--title`
  - `--comment`
- The **`--pr` help block** (flag line + indented continuations) documents multi-mode PR:
  - **show** / bare `--pr` (open PR URL or empty) — already present in product help
  - **status**: a status-with-pr hint such as literal `with --status` (PR metadata + workflow/check status)
  - **comment-only** / **create-attach** companions remain documented
- **Push + PR** rule is visible somewhere in help (soft wording; implementation-owned):
  - either the `--pr` block mentions `with --push` for tip push when an open PR exists / without title, **or**
  - the `--push` help documents `with --pr` (full-push tip then PR path / open-PR when applicable)
- Stdout (preferred for usage) ends with trailing `\n` when non-empty.

## Side Effects

- Read-only (`-h` only).

## Exit Code

- 0

```go
import (
	"regexp"
	"strings"

	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 for -h, got %d stdout=%q stderr=%q",
			resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	help := resp.Stdout + resp.Stderr
	if strings.TrimSpace(help) == "" {
		t.Fatal("expected non-empty help text for wrk -h")
	}
	if resp.Stdout != "" && !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("help stdout should end with trailing newline, got %q", resp.Stdout)
	}
	for _, want := range []string{"--pr", "--title", "--comment"} {
		if !strings.Contains(help, want) {
			t.Fatalf("help must mention %q; got stdout=%q stderr=%q", want, resp.Stdout, resp.Stderr)
		}
	}

	prBlock := helpFlagBlock(help, "--pr")
	if prBlock == "" {
		t.Fatalf("help must document --pr as a flag; full help:\n%s", help)
	}
	// Status-with-pr: require a clear --status companion on the --pr surface.
	// Prefer literal "with --status"; also accept "--status" inside the --pr block
	// (not the standalone git --status flag line).
	if !strings.Contains(prBlock, "with --status") && !prBlockMentionsStatusCompanion(prBlock) {
		t.Fatalf("--pr help block must document with --status (PR metadata + checks); prBlock=%q\nfull help:\n%s",
			prBlock, help)
	}

	// Push + PR: soft — wording lives under --pr and/or --push.
	pushBlock := helpFlagBlock(help, "--push")
	pushPrOK := strings.Contains(prBlock, "with --push") ||
		strings.Contains(pushBlock, "with --pr") ||
		(strings.Contains(pushBlock, "--pr") && strings.Contains(strings.ToLower(pushBlock), "push"))
	if !pushPrOK {
		t.Fatalf("help must document --push with --pr (tip full-push / open-PR path); prBlock=%q pushBlock=%q\nfull help:\n%s",
			prBlock, pushBlock, help)
	}
}

// helpFlagBlock returns the first flag line whose first token is flag plus any
// following indented continuation lines (lessflags multi-line descriptions).
func helpFlagBlock(help, flag string) string {
	re := regexp.MustCompile(`(?m)^  ` + regexp.QuoteMeta(flag) + `\b[^\n]*`)
	loc := re.FindStringIndex(help)
	if loc == nil {
		return ""
	}
	rest := help[loc[1]:]
	// Continuations are heavily indented (spaces after newline).
	var b strings.Builder
	b.WriteString(help[loc[0]:loc[1]])
	for {
		if !strings.HasPrefix(rest, "\n") {
			break
		}
		line := rest[1:]
		nl := strings.IndexByte(line, '\n')
		var one string
		if nl < 0 {
			one = line
			rest = ""
		} else {
			one = line[:nl]
			rest = line[nl:]
		}
		// Continuation lines start with many spaces and are not a new "  --flag" line.
		if strings.HasPrefix(one, "  --") {
			break
		}
		if strings.TrimSpace(one) == "" {
			// blank line ends the block
			break
		}
		// require indentation beyond two spaces
		if !strings.HasPrefix(one, "   ") {
			break
		}
		b.WriteByte('\n')
		b.WriteString(one)
		if rest == "" {
			break
		}
	}
	return b.String()
}

// prBlockMentionsStatusCompanion is a soft fallback when product avoids the
// exact phrase "with --status" but still documents --status under --pr.
func prBlockMentionsStatusCompanion(prBlock string) bool {
	if !strings.Contains(prBlock, "--status") {
		return false
	}
	lower := strings.ToLower(prBlock)
	// Prefer PR-status semantics (checks / reviews / metadata / workflow), not
	// a drive-by mention of the standalone git --status mode.
	for _, hint := range []string{"check", "review", "metadata", "workflow", "state", "pr status", "pull request"} {
		if strings.Contains(lower, hint) {
			return true
		}
	}
	// Bare "--status" inside the --pr multi-line block is still enough: the
	// standalone --status flag is a different help line, not nested here.
	return true
}
```
