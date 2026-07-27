
## Expected

- Exit code 0.
- Stdout includes scope plan line, `tagged v0.0.2 @ <short-hash>`, and `1 tag created`.
- Lightweight tag `v0.0.2` exists at HEAD.

## Exit Code

- 0

```go
import (

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"

	"fmt"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	short := shortHEAD(t, req.MainRepo)
	wantStdout := fmt.Sprintf(
		"v0.0.1        owned changed                  ->  v0.0.2\ntagged v0.0.2 @ %s\n1 tag created\n",
		short,
	)
	assert.Output(t, resp.Stdout, tagNextStdoutV2(wantStdout))

	if !tagRefExists(t, req.MainRepo, "v0.0.2") {
		t.Fatal("v0.0.2 tag should exist after apply")
	}
	got := gitOutputIsolated(t, req.MainRepo, "rev-parse", "v0.0.2")
	head := gitOutputIsolated(t, req.MainRepo, "rev-parse", "HEAD")
	if got != head {
		t.Fatalf("v0.0.2 should point at HEAD: tag=%s head=%s", got, head)
	}
}
```