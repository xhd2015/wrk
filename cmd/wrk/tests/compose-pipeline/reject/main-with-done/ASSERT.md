## Expected

- Non-zero exit.
- Stderr indicates mutual exclusion / not valid / cannot combine.
- Mentions `--main` and/or `--done` (generic “`--main` is mutually exclusive with other modes” is OK today).

## Errors

- `--main` cannot compose with `--done`.

## Exit Code

- Non-zero

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for --main --done; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	se := resp.Stderr
	if !strings.Contains(se, "mutually exclusive") &&
		!strings.Contains(se, "not valid") &&
		!strings.Contains(se, "cannot") {
		t.Fatalf("stderr should indicate exclusion, got %q", se)
	}
	// Accept generic --main mutex or a message that names --done.
	if !strings.Contains(se, "--main") &&
		!strings.Contains(strings.ToLower(se), "main") &&
		!strings.Contains(se, "--done") &&
		!strings.Contains(se, "done") {
		t.Fatalf("stderr should mention --main and/or --done, got %q", se)
	}
}
```
