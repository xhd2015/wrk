## Expected

- Exit code 0.
- Stdout is the embedded `SKILL.md` from the wrk binary.
- Stdout contains `WRK_SKILL_DOCTEST_MARKER` and YAML `name: wrk`.
- Stdout ends with trailing `\n`.
- Stderr is empty.

## Side Effects

- Read-only.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmbeddedSkillStdout(t, resp.Stdout)
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
}
```
