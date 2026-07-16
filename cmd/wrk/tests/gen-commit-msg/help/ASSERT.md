## Expected

- Exit code 0.
- Help text (stdout and/or stderr) documents gen-commit-msg usage tokens:
  - mode identity: `--gen-commit-msg` and/or `gen-commit-msg`
  - flags: `--model`, `--dry-run`, `--commit`, `--no-verify`, `--agent-runner`
- Stdout (preferred for usage) ends with trailing `\n` when non-empty.

## Side Effects

- Read-only (`-h` only).

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	help := resp.Stdout + resp.Stderr
	if strings.TrimSpace(help) == "" {
		t.Fatal("expected non-empty help text for wrk --gen-commit-msg -h")
	}
	if resp.Stdout != "" && !strings.HasSuffix(resp.Stdout, "\n") {
		t.Fatalf("help stdout should end with trailing newline, got %q", resp.Stdout)
	}
	// Mode identity: flag form preferred; bare name from library usage is acceptable.
	if !strings.Contains(help, "--gen-commit-msg") && !strings.Contains(help, "gen-commit-msg") {
		t.Fatalf("help must mention gen-commit-msg / --gen-commit-msg; got %q", help)
	}
	for _, want := range []string{"--model", "--dry-run", "--commit", "--no-verify", "--agent-runner"} {
		if !strings.Contains(help, want) {
			t.Fatalf("help must mention %q; got stdout=%q stderr=%q", want, resp.Stdout, resp.Stderr)
		}
	}
	assert.Output(t, help, `<contains>
--dry-run
</contains>`)
}
```
