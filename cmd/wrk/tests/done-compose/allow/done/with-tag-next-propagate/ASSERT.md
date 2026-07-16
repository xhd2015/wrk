## Expected

- Flag layer accepts `--done` + `--tag-next` + `--propagate-tags` (no mutual-exclusion error).
- Stderr must **not** contain `mutually exclusive`.
- Exit may be non-zero for later-stage reasons on a main-repo cwd.

## Side Effects

- None required (flag-layer only).

## Exit Code

- Any (0 or non-zero), as long as the failure is not flag mutual exclusion.

```go
import (
	"strings"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	se := resp.Stderr
	if strings.Contains(se, "mutually exclusive") {
		t.Fatalf("flag layer still rejects --done --tag-next --propagate-tags as mutually exclusive; stderr=%q stdout=%q exit=%d",
			se, resp.Stdout, resp.ExitCode)
	}
}
```
