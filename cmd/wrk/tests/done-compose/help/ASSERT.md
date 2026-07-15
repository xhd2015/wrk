## Expected

- Exit code 0.
- Help text (stdout and/or stderr) documents composition for finish modes:
  - The **`--done`** flag synopsis line includes optional **`--tag-next`** and **`--push`**
    (and still documents optional **`--sync`**).
  - The **`--merge-back`** flag synopsis line similarly includes optional **`--tag-next`**
    and **`--push`** (composition parity with `--done`).
  - **`--push`** is not documented only as “with `--tag-next`”: help must also indicate
    validity with a primary (`--done` and/or `--merge-back`).
  - Help must **not** claim that `--tag-next` is mutually exclusive with `--done`.
  - **`--json`** remains tagged as for bare `--tag-next` (not a primary composition flag);
    help must not claim `--json` is valid with `--done`.
- Prefer implementer wording (soft, not asserted verbatim):
  - `--done [--sync] [--tag-next] [--push] [--dry-run] …`
  - `--push` dual meaning: tags with `--tag-next`; branch (and tags when combined) with primary.

## Side Effects

- Read-only (`--help` only).

## Exit Code

- 0

```go
import (
	"regexp"
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 for --help, got %d stdout=%q stderr=%q", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	help := resp.Stdout + resp.Stderr
	if strings.TrimSpace(help) == "" {
		t.Fatal("expected non-empty help text")
	}

	// --done synopsis must list composition post-modifiers on the same flag line.
	doneLine := firstFlagLine(help, "--done")
	if doneLine == "" {
		t.Fatalf("help must document --done; got %q", help)
	}
	for _, flag := range []string{"--sync", "--tag-next", "--push"} {
		if !strings.Contains(doneLine, flag) {
			t.Fatalf("--done help line must mention optional %s for composition; line=%q\nfull help:\n%s", flag, doneLine, help)
		}
	}

	// --merge-back synopsis parity (same post-modifier composition).
	mergeLine := firstFlagLine(help, "--merge-back")
	if mergeLine == "" {
		t.Fatalf("help must document --merge-back; got %q", help)
	}
	for _, flag := range []string{"--tag-next", "--push"} {
		if !strings.Contains(mergeLine, flag) {
			t.Fatalf("--merge-back help line must mention optional %s for composition; line=%q\nfull help:\n%s", flag, mergeLine, help)
		}
	}

	// --push dual meaning: not only "with --tag-next".
	pushLine := firstFlagLine(help, "--push")
	if pushLine == "" {
		t.Fatalf("help must document --push; got %q", help)
	}
	// Prefer the flag definition line; also allow a short multi-line description.
	pushBlob := pushLine + "\n" + nextIndentedContinuation(help, pushLine)
	mentionsTagNext := strings.Contains(pushBlob, "--tag-next")
	mentionsPrimary := strings.Contains(pushBlob, "--done") ||
		strings.Contains(pushBlob, "--merge-back") ||
		strings.Contains(strings.ToLower(pushBlob), "primary")
	if !mentionsPrimary {
		t.Fatalf("--push help must document validity with --done/--merge-back (dual meaning), not only with --tag-next; blob=%q\nfull help:\n%s", pushBlob, help)
	}
	if !mentionsTagNext {
		// Soft preference: keep tag-next meaning too; still require primary above.
		// Fail only if help lost tag-next entirely for push context.
		if !strings.Contains(help, "--tag-next") {
			t.Fatalf("help must still mention --tag-next (push dual meaning includes tags); got %q", help)
		}
	}

	// No exclusivity claim for tag-next vs done in help text.
	lower := strings.ToLower(help)
	forbidden := []string{
		"tag-next is mutually exclusive with --done",
		"--tag-next is mutually exclusive with --done",
		"--done is mutually exclusive with --tag-next",
		"tag-next is only valid alone",
		"--tag-next cannot be combined with --done",
	}
	for _, phrase := range forbidden {
		if strings.Contains(lower, strings.ToLower(phrase)) {
			t.Fatalf("help must not claim tag-next exclusive with done; found %q in help", phrase)
		}
	}

	// --json is not a primary composition flag (docs should not invite --done --json).
	jsonLine := firstFlagLine(help, "--json")
	if jsonLine != "" {
		if strings.Contains(jsonLine, "--done") || strings.Contains(jsonLine, "--merge-back") {
			// Allowed only if the line is rejecting those; otherwise fail.
			if !strings.Contains(strings.ToLower(jsonLine), "not") &&
				!strings.Contains(strings.ToLower(jsonLine), "only") {
				t.Fatalf("--json help must not present validity with primary modes; line=%q", jsonLine)
			}
		}
	}
}

// firstFlagLine returns the first help line whose first flag token is flag
// (e.g. "  --done …" or "  --push …"), ignoring bracketed mentions like [--push].
func firstFlagLine(help, flag string) string {
	re := regexp.MustCompile(`(?m)^[ \t]*` + regexp.QuoteMeta(flag) + `\b.*$`)
	m := re.FindString(help)
	return strings.TrimRight(m, "\r")
}

// nextIndentedContinuation returns the line immediately after flagLine when it is
// a continuation indent (common for multi-line flag descriptions).
func nextIndentedContinuation(help, flagLine string) string {
	idx := strings.Index(help, flagLine)
	if idx < 0 {
		return ""
	}
	rest := help[idx+len(flagLine):]
	rest = strings.TrimPrefix(rest, "\r")
	rest = strings.TrimPrefix(rest, "\n")
	if rest == "" {
		return ""
	}
	end := strings.IndexByte(rest, '\n')
	var line string
	if end < 0 {
		line = rest
	} else {
		line = rest[:end]
	}
	// Continuation lines under Flags: are usually indented past the flag column.
	if strings.HasPrefix(line, "  ") || strings.HasPrefix(line, "\t") {
		if strings.TrimSpace(line) == "" {
			return ""
		}
		// Stop if next line is another top-level flag.
		if regexp.MustCompile(`^[ \t]*--[a-zA-Z]`).MatchString(line) {
			return ""
		}
		return line
	}
	return ""
}
```
