---
label: slow
---
## Expected

- Exit code 0.
- Stdout contains two detailed status blocks (absolute `Dir`, `Remote`, `Worktrees`) in lexicographic order with a blank line separator.
- `projects.json` still holds exactly two entries.

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
	assertProjectsCount(t, req.WrkHome, 2)

	paths := []string{resolvePath(t, req.MainRepo), resolvePath(t, req.SecondRepo)}
	sort.Strings(paths)
	repoByPath := map[string]string{
		resolvePath(t, req.MainRepo):  req.MainRepo,
		resolvePath(t, req.SecondRepo): req.SecondRepo,
	}

	var blocks []string
	for _, abs := range paths {
		repo := repoByPath[abs]
		block := projectListBlock(t, repo)
		blocks = append(blocks, block)
	}
	assert.Output(t, resp.Stdout, projectsStdoutV2(t, blocks...))
}
```