# Scenario

**Feature**: failed create with WRK_FOLLOWUP_FILE writes no follow-up

```
plain non-git dir + WRK_FOLLOWUP_FILE=tmp
wrk -> non-zero; follow-up file empty
```

## Steps

1. Create non-git directory.
2. Run `wrk --new` with follow-up env set (create fails on non-git).

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
	req.UseFollowupEnv = true
	req.CLIArgs = []string{"--new"}
	return nil
}
```
