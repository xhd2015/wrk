## Expected

- Exit code 0.
- Stdout is the embedded `SKILL.md` (marker + `name: wrk`).
- Stdout documents `--propagate-tags` (consumer require bump to source releases).
- Stdout documents `--projects-dep-graph` (cross-project module dep graph).
- Stdout documents `--pr` as its own flag token (not only as a prefix of `--propagate-tags`).
- Stdout documents `--title` and `--comment` as required companions of `--pr`.
- Stderr empty.

## Side Effects

- Read-only.

## Exit Code

- 0

```go
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
}
```
