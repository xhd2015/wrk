## Expected

- Non-zero exit; empty stdout.
- Stderr contains `unexpected arguments` (or equivalent `wrk: unexpected arguments`).

## Errors

- `--main` takes no path positional / basename.

## Exit Code

- Non-zero

```go
import "github.com/xhd2015/doctest/assert"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero exit, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assertEmptyStdout(t, resp.Stdout)
	assert.Output(t, resp.Stderr, `<contains>
unexpected arguments
</contains>`)
}
```
