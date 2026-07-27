
## Expected Output

```
v0.0.1        same commit                    ->  skip
0 tag planned
```

## Expected

- Exit code 0.
- Stdout shows skip with `0 tag planned`.
- No new tags created.

## Exit Code

- 0

```go
import (
	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q stdout=%q", resp.ExitCode, resp.Stderr, resp.Stdout)
	}

	wantStdout := "v0.0.1        same commit                    ->  skip\n0 tag planned\n"
	assert.Output(t, resp.Stdout, tagNextStdoutV2(wantStdout))

	if tagRefExists(t, req.MainRepo, "v0.0.2") {
		t.Fatal("no tag should be created when nothing changed")
	}
}
```