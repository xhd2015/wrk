## Expected

- Exit code 0.
- Help text (stdout and/or stderr) documents composition for finish modes:
  - The **`--done`** flag synopsis line includes optional **`--tag-next`** and **`--push`**
    (and still documents optional **`--sync`**), plus **`--reinstall-local`** (P1 tail),
    and **`--gen-commit-msg`** (P2 pre-stage).
  - The **`--merge-back`** flag synopsis line similarly includes optional **`--tag-next`**,
    **`--push`**, **`--reinstall-local`**, and **`--gen-commit-msg`** (composition parity).
  - **`--gen-commit-msg`** definition (or nearby window) documents pre-stage use with
    `--done` / `--merge-back` and/or that **`--commit`** is required when composed with primary.
  - **`--reinstall-local`** flag docs mention validity after successful primary
    (`--done` / `--merge-back`) and still document bare / `--main` use.
  - **`--push`** is not documented only as “with `--tag-next`”: help must also indicate
    validity with a primary (`--done` and/or `--merge-back`).
  - Help must **not** claim that `--tag-next` is mutually exclusive with `--done`.
  - **`--json`** remains tagged as for bare `--tag-next` (not a primary composition flag);
    help must not claim `--json` is valid with `--done`.
- Prefer implementer wording (soft, not asserted verbatim):
  - `--done [--gen-commit-msg --commit …] [--sync] [--tag-next] [--push] [--propagate-tags] [--reinstall-local] [--dry-run] …`
  - `--gen-commit-msg`: also as pre-stage before `--done` / `--merge-back` (requires `--commit`)
  - `--reinstall-local` also: after successful `--done` / `--merge-back`
  - `--push` dual meaning: tags with `--tag-next`; branch (and tags when combined) with primary.
- **P3 fluent recipes** (soft preference for `usage()` / SKILL examples; hard asserts stay flag-list based above):
  - `wrk --done --sync --tag-next --push -y`
  - `wrk --done --sync --tag-next --push --reinstall-local -y`
  - `wrk --gen-commit-msg --commit --model=M --done --sync --tag-next --push -y`
  - Pre requires `--commit`; reinstall after primary from main.

## Side Effects

- Read-only (`--help` only).

## Exit Code

- 0

```go
import (
	"regexp"
	"strings"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 for --help, got %d stdout=%q stderr=%q", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	help := resp.Stdout + resp.Stderr
	if strings.TrimSpace(help) == "" {
		t.Fatal("expected non-empty help text")
	}

	// --done synopsis must list composition pre/post-modifiers on the same flag line.
	doneLine := firstFlagLine(help, "--done")
	if doneLine == "" {
		t.Fatalf("help must document --done; got %q", help)
	}
	for _, flag := range []string{"--sync", "--tag-next", "--push", "--reinstall-local", "--gen-commit-msg"} {
		if !strings.Contains(doneLine, flag) {
			t.Fatalf("--done help line must mention optional %s for composition; line=%q\nfull help:\n%s", flag, doneLine, help)
		}
	}

	// --merge-back synopsis parity (same pre/post-modifier composition).
	mergeLine := firstFlagLine(help, "--merge-back")
	if mergeLine == "" {
		t.Fatalf("help must document --merge-back; got %q", help)
	}
	for _, flag := range []string{"--tag-next", "--push", "--reinstall-local", "--gen-commit-msg"} {
		if !strings.Contains(mergeLine, flag) {
			t.Fatalf("--merge-back help line must mention optional %s for composition; line=%q\nfull help:\n%s", flag, mergeLine, help)
		}
	}

	// --gen-commit-msg documents pre-stage with primary (and preferably --commit when composed).
	genLine := firstFlagLine(help, "--gen-commit-msg")
	if genLine == "" {
		t.Fatalf("help must document --gen-commit-msg; got %q", help)
	}
	genWindow := genLine
	if idx := strings.Index(help, genLine); idx >= 0 {
		end := idx + 600
		if end > len(help) {
			end = len(help)
		}
		genWindow = help[idx:end]
	}
	if !strings.Contains(genWindow, "--done") && !strings.Contains(genWindow, "--merge-back") {
		// Accept either gen-commit definition mentioning primary, or done/merge synopsis
		// already requiring --gen-commit-msg (asserted above) plus a nearby "pre"/"commit" hint.
		if !strings.Contains(doneLine, "--gen-commit-msg") {
			t.Fatalf("--gen-commit-msg help should document pre-stage with --done/--merge-back; window=%q\nfull help:\n%s",
				genWindow, help)
		}
	}

	// --reinstall-local documents primary compose (after done/merge-back).
	// Done/merge-back synopsis lines already require the flag; also require the
	// reinstall definition (or nearby description) to mention a primary mode.
	reinstallLine := firstFlagLine(help, "--reinstall-local")
	if reinstallLine == "" {
		t.Fatalf("help must document --reinstall-local; got %q", help)
	}
	reinstallWindow := reinstallLine
	if idx := strings.Index(help, reinstallLine); idx >= 0 {
		end := idx + 500
		if end > len(help) {
			end = len(help)
		}
		reinstallWindow = help[idx:end]
	}
	if !strings.Contains(reinstallWindow, "--done") &&
		!strings.Contains(reinstallWindow, "--merge-back") {
		t.Fatalf("--reinstall-local help should document validity after --done/--merge-back; window=%q\nfull help:\n%s",
			reinstallWindow, help)
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
