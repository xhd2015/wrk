## Expected

- Non-zero exit (dirty cascade target / preflight).
- Stderr contains `Error:` with **red** ANSI CSI (`\x1b[31m`) coloring the prefix (D9).
- Prefer prefix-only; `colorize("Error:", red)` or red CSI immediately before `Error:` OK.
- Today Error: is plain (no color) → **RED** until shared Error: print path colors.

## Exit Code

- Non-zero

```go
import "strings"

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("force-color Error: leaf needs hard error (dirty cascade); got exit 0 stdout=%q stderr=%q",
			resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(resp.Stderr, "Error:") {
		t.Fatalf("stderr must include Error:; stderr:\n%s\nstdout:\n%s", resp.Stderr, resp.Stdout)
	}
	assertStderrHasRedErrorPrefix(t, resp.Stderr)
}
```
