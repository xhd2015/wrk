## Expected Output

```
==== dep-replace (dry-run) ====
dep  example.com/dep => <abs>

  checkout  .
    module  example.com/consumer
      would: replace  example.com/dep => <abs>

dep-replace: would replace in 1 modules in 1 checkouts
```

## Expected

- Exit 0.
- Dry-run banner + `would: replace` (not `no such dir: …/--dry-run`).
- go.mod unchanged.

## Side Effects

- None.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	if strings.Contains(resp.Stderr, "--dry-run") && strings.Contains(resp.Stderr, "no such dir") {
		t.Fatalf("--dry-run must not be parsed as a directory; stderr=%q", resp.Stderr)
	}
	assert.Output(t, resp.Stdout, `---
version: 3
__ABS__: type=string
---
==== dep-replace \(dry-run\) ====
dep  example\.com/dep => __ABS__

  checkout  \.
    module  example\.com/consumer
      would: replace  example\.com/dep => __ABS__

dep-replace: would replace in 1 modules in 1 checkouts
`)
	assertGoModUnchanged(t, req)
	assertNoTidyArtifacts(t, req)
}
```
