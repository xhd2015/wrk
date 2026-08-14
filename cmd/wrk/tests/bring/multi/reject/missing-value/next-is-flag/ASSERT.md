## Expected

- Non-zero exit.
- Stderr matches library wording: `--bring` **requires a value** (flag is not a value).
- Must not succeed as `--no-dep` without a bring path.
- No `external/` under WorkRoot.

## Errors

- `--bring` immediately followed by another flag.

## Exit Code

- Non-zero

```go
import (
	"path/filepath"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	if resp.ExitCode == 0 {
		t.Fatalf("expected non-zero for --bring --no-dep, got 0 stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assert.Output(t, resp.Stderr, `---
version: 3
---
.*requires a value.*
`)
	assertFileNotExists(t, filepath.Join(req.WorkRoot, "external"))
}
```
