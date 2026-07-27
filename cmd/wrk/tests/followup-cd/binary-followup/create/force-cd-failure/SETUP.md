# Scenario

**Feature**: failed create with --force-cd does not land (no follow-up, no shell)

```
plain non-git dir + WRK_FOLLOWUP_FILE + fake bash
wrk --force-cd -> non-zero; follow-up empty; fake shell not launched
```

## Steps

1. Create non-git directory; install fake bash to detect accidental shell.
2. Run `wrk --force-cd` with follow-up env set.

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
	req.CLIArgs = []string{"--force-cd"}
	installFakeBash(t, req, 0)
	return nil
}
```
