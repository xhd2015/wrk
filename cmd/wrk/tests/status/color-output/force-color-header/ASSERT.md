## Expected Output

```text
Dir:          .
Branch:       main
Commit:       <hash>  root status repo
Status:       <green>clean</green>
Remote:       (no upstream)

<gray>---- external ----</gray>

Dir:          tools/child
Branch:       main
Commit:       <hash>  child status repo
Status:       <green>clean</green>
```

## Expected

- Exit code 0; stderr empty.
- Exactly two `Dir:` blocks (main primary, then nested external).
- Header line is full-line gray ANSI around exactly `---- external ----` (no extra spaces).
- Nested block still appears after the header.
- Other block coloring unchanged (green `clean` on both Status values).

## Exit Code

- 0

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	if got := statusOutputBlockCount(resp.Stdout); got != 2 {
		t.Fatalf("expected 2 status blocks, got %d:\n%s", got, resp.Stdout)
	}

	assert.Output(t, resp.Stdout, statusStdoutPrimaryExternalColored(t,
		[]string{
			colorStatusBlockPlain(t, req.MainRepo, ".", "<ansi-color green>clean</ansi-color>", ""),
		},
		[]string{
			colorStatusBlockPlain(t, req.DepPath, "tools/child", "<ansi-color green>clean</ansi-color>", ""),
		},
	))
}
```
