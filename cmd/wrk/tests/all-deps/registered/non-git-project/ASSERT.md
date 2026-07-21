## Expected

- Exit code 0.
- Stdout is exactly `wrked 0 deps\n`.
- No `external/` directory; no `replace` directives added.

## Exit Code

- 0

```go
import "github.com/xhd2015/doctest/assert"

import (
	"path/filepath"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}
	assert.Output(t, resp.Stdout, allDepsStdoutV2("wrked 0 deps\n"))
	mod, err := allDepsReadGoMod(req.RepoDir)
	if err != nil {
		t.Fatalf("read go.mod: %v", err)
	}
	if len(mod.Replace) != 0 {
		t.Fatalf("go.mod should have no replace directives, got %+v", mod.Replace)
	}
	assertFileNotExists(t, filepath.Join(req.ConsumerTop, "external"))
}
```