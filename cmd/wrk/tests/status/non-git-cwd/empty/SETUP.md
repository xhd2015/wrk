# Scenario

**Feature**: wrk --status on an empty plain directory succeeds with no blocks

```
# no .git ancestor and no nested git checkouts
plain empty cwd -> wrk --status -> exit 0, empty stdout
```

## Steps

1. Create `{WorkRoot}/plain` without a `.git` directory and without nested repos.
2. Run `wrk --status` from that directory.

```go
import (
	"path/filepath"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	plain := filepath.Join(req.WorkRoot, "plain")
	mkdirAll(t, plain)

	req.RepoDir = plain
	return nil
}
```
