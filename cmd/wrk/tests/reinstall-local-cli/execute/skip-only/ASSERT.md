
## Expected Output

```
skip: missing (not in <BinDir>)
reinstalled 0, skipped 1, failed 0
```

## Expected

- Exit code 0 (successful execute with zero install attempts).
- Stdout: one `skip:` line for `missing` naming resolved GOBIN path, then summary
  `reinstalled 0, skipped 1, failed 0\n`.
- No `go install` / `go run` progress lines and no `would:` dry-run vocabulary.
- GOBIN remains empty of a `missing` binary.

## Side Effects

- No `go install` / `go run` of candidates; no new bin under GOBIN.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/session"
	"fmt"
	"path/filepath"
	"os"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	_ = d
	assertErrIsNil(t, err)
	assertExitZero(t, resp)
	want := fmt.Sprintf("skip: missing (not in %s)\nreinstalled 0, skipped 1, failed 0\n", req.BinDir)
	assertOutputExact(t, resp.Stdout, v2StdoutTemplate(want))
	assertNotContains(t, resp.Stdout, "go install")
	assertNotContains(t, resp.Stdout, "go run")
	assertNotContains(t, resp.Stdout, "would:")
	if _, statErr := os.Stat(filepath.Join(req.BinDir, "missing")); !os.IsNotExist(statErr) {
		t.Fatalf("GOBIN must not contain missing bin; stat err=%v", statErr)
	}
}
```
