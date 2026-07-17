## Expected

- Non-zero exit.
- Stderr mentions `--json`.
- Indicates not valid / only bare tag-next / cannot combine with other stages.
- Prefer naming a co-present stage (`--sync` or `--tag-next` multi-stage context).

## Errors

- `--json` is not valid with multi-stage compose.

## Exit Code

- Non-zero

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for multi-stage --json; stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	se := resp.Stderr
	if !strings.Contains(se, "--json") && !strings.Contains(se, "json") {
		t.Fatalf("stderr should mention --json, got %q", se)
	}
	if !strings.Contains(se, "not valid") &&
		!strings.Contains(se, "mutually exclusive") &&
		!strings.Contains(se, "cannot") &&
		!strings.Contains(se, "only valid") &&
		!strings.Contains(se, "only with") {
		t.Fatalf("stderr should reject multi-stage+json policy, got %q", se)
	}
}
```
