## Expected

- Exit code 0.
- Stdout is the embedded `SKILL.md` (marker + `name: wrk`).
- Stdout documents `--propagate-tags` (consumer require bump to source releases).
- Stdout documents `--projects-dep-graph` (cross-project module dep graph).
- Stdout documents `--pr` as its own flag token (not only as a prefix of `--propagate-tags`).
- Stdout documents `--title` and `--comment` as **create/attach** companions of `--pr`
  (tokens present); they are **not** framed as always-required for every `--pr` mode.
- Multi-mode PR surface (implementation-owned wording; soft checks):
  - status-with-pr (e.g. `--pr --status` or status under the Pull request section)
  - bare/show `--pr` and/or comment-only / create compose examples remain agent-usable
- Stderr empty.

## Side Effects

- Read-only.

## Exit Code

- 0

```go
import "strings"

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmbeddedSkillStdout(t, resp.Stdout)
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	// Token-aware: "--pr" must not match only via "--propagate-tags".
	for _, flag := range []string{
		"--propagate-tags",
		"--projects-dep-graph",
		"--pr",
		"--title",
		"--comment",
	} {
		if !skillDocumentsFlagToken(resp.Stdout, flag) {
			t.Fatalf("embedded SKILL.md should document %s as a flag token, stdout:\n%s", flag, resp.Stdout)
		}
	}

	// Outdated single-mode wording: title+comment must not be claimed as always
	// required companions of every --pr invocation (multi-mode contract).
	out := resp.Stdout
	outdated := []string{
		"required companions of `--pr`",
		"required companions of --pr",
		"are required companions of `--pr`",
		"are required companions of --pr",
	}
	for _, phrase := range outdated {
		if strings.Contains(out, phrase) {
			t.Fatalf("embedded SKILL.md must not claim title+comment always required with --pr (multi-mode); found %q\n%s",
				phrase, out)
		}
	}

	// Multi-mode: status-with-pr must appear in the Pull request section (or
	// adjacent multi-mode examples). Soft: accept several phrasings.
	prSection := skillPullRequestSection(out)
	statusOK := strings.Contains(prSection, "--pr --status") ||
		strings.Contains(out, "wrk --pr --status") ||
		strings.Contains(prSection, "with `--status`") ||
		strings.Contains(prSection, "with --status") ||
		(strings.Contains(prSection, "--status") && strings.Contains(prSection, "--pr"))
	if !statusOK {
		t.Fatalf("embedded SKILL.md Pull request section must document multi-mode --pr --status; section=%q\nfull:\n%s",
			prSection, out)
	}
}

// skillPullRequestSection returns the "## Pull request" body through the next
// top-level "## " heading, or the whole skill text if the heading is missing
// (so asserts still fail with a useful dump).
func skillPullRequestSection(stdout string) string {
	const h = "## Pull request"
	i := strings.Index(stdout, h)
	if i < 0 {
		// fallback: any line containing "Pull request" as a heading-ish marker
		for _, alt := range []string{"## Pull Request", "# Pull request", "Pull request"} {
			if j := strings.Index(stdout, alt); j >= 0 {
				i = j
				break
			}
		}
	}
	if i < 0 {
		return stdout
	}
	rest := stdout[i:]
	// skip first heading line
	if nl := strings.IndexByte(rest, '\n'); nl >= 0 {
		rest = rest[nl+1:]
	}
	// next top-level ## heading ends the section
	if j := strings.Index(rest, "\n## "); j >= 0 {
		return strings.TrimSpace(rest[:j])
	}
	return strings.TrimSpace(rest)
}
```
