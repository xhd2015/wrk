# Scenario

**Feature**: non-git cwd yields not a git repository

```
WorkRoot/plain/ (no .git)
  -> wrk --pin-locals
  -> non-zero
  -> stderr: not a git repository
```

## Steps

1. Create plain directory (no git init).
2. Run `wrk --pin-locals` with RepoDir = plain.

```go
import (
	"os"
	"path/filepath"

	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	plain := filepath.Join(req.WorkRoot, "plain")
	if err := os.MkdirAll(plain, 0o755); err != nil {
		return err
	}
	req.RepoDir = plain
	req.Args = []string{"--pin-locals"}
	return nil
}
```
