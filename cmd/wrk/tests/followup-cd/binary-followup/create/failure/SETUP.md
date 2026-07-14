# Scenario

**Feature**: failed create with WRK_FOLLOWUP_FILE writes no follow-up

```
plain non-git dir + WRK_FOLLOWUP_FILE=tmp
wrk -> non-zero; follow-up file empty
```

## Steps

1. Create non-git directory.
2. Run bare `wrk` with follow-up env set.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	plain := filepath.Join(req.WorkRoot, "plain")
	mkdirAll(t, plain)
	req.RepoDir = plain
	req.UseFollowupEnv = true
	req.CLIArgs = nil
	return nil
}
```
