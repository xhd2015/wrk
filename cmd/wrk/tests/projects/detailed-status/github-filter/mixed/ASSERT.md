## Expected

- Exit code 0.
- Stdout is one status block for the github.com project only.
- `projects.json` still has three entries.
- Stderr is empty.

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
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	assertProjectsCount(t, req.WrkHome, 3)
	assertProjectsBlocksSeparated(t, resp.Stdout, 1)

	// After set-url to github.com, upstream tracking still uses origin/main.
	remote := compareWithRemoteField(t, req.MainRepo, "origin/main", "main")
	block := projectStatusBlockTemplate(t, req.MainRepo, "clean", remote, "0 total, 0 dirty")
	assert.Output(t, resp.Stdout, block)

	localDir := projectDirLine(t, req.SecondRepo)
	noRemoteDir := projectDirLine(t, req.DepPath)
	if strings.Contains(resp.Stdout, localDir) {
		t.Fatalf("stdout should omit local-origin project, contains %q", localDir)
	}
	if strings.Contains(resp.Stdout, noRemoteDir) {
		t.Fatalf("stdout should omit no-remote project, contains %q", noRemoteDir)
	}
}
```
