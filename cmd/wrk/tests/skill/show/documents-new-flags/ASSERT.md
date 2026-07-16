## Expected

- Exit code 0.
- Stdout is the embedded `SKILL.md` (marker + `name: wrk`).
- Stdout documents `--propagate-tags` (consumer require bump to source releases).
- Stdout documents `--projects-dep-graph` (cross-project module dep graph).
- Stderr empty.

## Side Effects

- Read-only.

## Exit Code

- 0

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assertEmbeddedSkillStdout(t, resp.Stdout)
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	for _, flag := range []string{"--propagate-tags", "--projects-dep-graph"} {
		if !strings.Contains(resp.Stdout, flag) {
			t.Fatalf("embedded SKILL.md should document %s, stdout:\n%s", flag, resp.Stdout)
		}
	}
}
```

