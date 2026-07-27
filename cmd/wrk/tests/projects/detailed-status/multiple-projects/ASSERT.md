---
label: slow
---
## Expected

- Exit code 0.
- Two detailed status blocks in lexicographic order (`aaa` before `zzz`).
- Blank line between blocks.
- Stderr is empty.

## Exit Code

- 0

```go
import (
	"sort"

	"github.com/xhd2015/doctest/assert"
	"github.com/xhd2015/doctest/session"
)

func Assert(t *testing.T, d *session.Doctest, req *Request, resp *Response, err error) {
	assertErrIsNil(t, err)
	if resp.ExitCode != 0 {
		t.Fatalf("exit code %d stderr=%q", resp.ExitCode, resp.Stderr)
	}
	if resp.Stderr != "" {
		t.Fatalf("stderr should be empty, got %q", resp.Stderr)
	}
	assertProjectsBlocksSeparated(t, resp.Stdout, 2)

	paths := []string{resolvePath(t, req.MainRepo), resolvePath(t, req.SecondRepo)}
	sort.Strings(paths)
	repoByPath := map[string]string{
		resolvePath(t, req.MainRepo):  req.MainRepo,
		resolvePath(t, req.SecondRepo): req.SecondRepo,
	}

	var blocks []string
	for _, abs := range paths {
		repo := repoByPath[abs]
		remote := compareWithRemoteField(t, repo, "origin/main", "main")
		blocks = append(blocks, projectStatusBlockExact(t, repo, "clean", remote, "0 total, 0 dirty"))
	}
	assert.Output(t, resp.Stdout, projectsStdoutV2(t, blocks...))
}
```