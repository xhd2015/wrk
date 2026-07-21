## Expected Output

Stderr plans the add-all step:

```text
<contains>
would: git add -A
</contains>
```

## Expected

- Exit code 0.
- Stdout is mock message B for N=1.
- Stderr contains the dry-run plan line `would: git add -A`.
- Stderr does **not** log a real `$ git add -A` execution line.
- Combined output does **not** report wrk/library "unrecognized flag" / "unknown flag" for `--add-all`.

## Side Effects

- Agent is not required (dry-run pure plan).
- Index need not grow (dry-run does not mutate); staged `change.go` may remain staged.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)

	combined := resp.Stdout + "\n" + resp.Stderr
	lower := strings.ToLower(combined)
	if strings.Contains(lower, "unrecognized flag") || strings.Contains(lower, "unknown flag") {
		t.Fatalf("wrk/--gen-commit-msg must accept --add-all (not unrecognized/unknown flag), stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 for --add-all --dry-run, got %d stdout=%q stderr=%q",
			resp.ExitCode, resp.Stdout, resp.Stderr)
	}

	assertMockMessageB(t, resp.Stdout, 1)

	if !strings.Contains(resp.Stderr, "would: git add -A") {
		t.Fatalf("stderr must contain would: git add -A, got stderr:\n%s\nstdout:\n%s",
			resp.Stderr, resp.Stdout)
	}
	if strings.Contains(resp.Stderr, "$ git add -A") {
		t.Fatalf("dry-run must not log real $ git add -A, stderr:\n%s", resp.Stderr)
	}

	assert.Output(t, resp.Stderr, `<contains>
would: git add -A
</contains>`)
}
```
