# Scenario

**Feature**: wrapper returns non-zero when follow-up cd fails (binary exited 0)

```
fake wrk exits 0 and writes "cd /no/such/path"
source bash.sh; wrk
  -> stderr contains cd line; exit != 0; FinalPWD unchanged
```

## Steps

1. Use fake `wrk` on PATH that writes a bad follow-up and exits 0.
2. Source installed bash.sh and invoke `wrk` from WorkRoot.

```go
import (
	"path/filepath"
)

func Setup(t *testing.T, req *Request) error {
	req.RepoDir = req.WorkRoot
	req.StartDir = req.WorkRoot
	req.UseFakeWrk = true
	req.FakeFollowupCD = filepath.Join(req.WorkRoot, "definitely-missing-followup-target")
	req.CLIArgs = nil
	return nil
}
```
