## Expected

- Exit code 0.
- Help text (stdout and/or stderr) contains `--scan-git-repos`.
- Help text also contains `--no-cache`.

## Exit Code

- 0

```go
import (
	"strings"

	"github.com/xhd2015/doctest/assert"
)

func Assert(t *testing.T, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("expected exit 0 for -h, got %d stdout=%q stderr=%q", resp.ExitCode, resp.Stdout, resp.Stderr)
	}
	help := resp.Stdout + resp.Stderr
	if !strings.Contains(help, "--scan-git-repos") {
		t.Fatalf("help must mention --scan-git-repos; got stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	if !strings.Contains(help, "--no-cache") {
		t.Fatalf("help must mention --no-cache; got stdout=%q stderr=%q", resp.Stdout, resp.Stderr)
	}
	assert.Output(t, help, `<contains>
--scan-git-repos
</contains>`)
}
```
