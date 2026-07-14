## Expected Output

```
---
name: wrk
description: >-
  Git worktree helper for isolated feature branches. Use when creating
  linked worktrees, merging back, checking status, linking deps, or
  looking up registered projects by basename.
---
```

## Expected

- Exit code 0.
- Stdout is the YAML frontmatter block with `---` delimiters; no body lines.
- Stdout ends with trailing `\n`.
- Stderr is empty.

## Exit Code

- 0

```go
func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertOutputExact(t, resp.Stdout, skillHeaderStdoutV2())
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
}
```
