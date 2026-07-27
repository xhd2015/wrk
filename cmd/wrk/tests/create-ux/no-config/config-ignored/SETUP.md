# Scenario

**Feature**: full create.* config is ignored when `--no-config` is set (no UX flags)

```
# config window+terminal+agent all on; no CLI UX flags
wrk --no-config
  -> exit 0; stdout = abs worktree path\n under WRK_HOME
  -> no space.CreateAndActivate
  -> no iterm OpenConfig
  -> no agent-run (outer or follow-up)

# contrast: same config without --no-config would invoke full UX
# (see pipeline/config/defaults-match-flags)
```

## Steps

1. Write full create UX config (`writeFullCreateUXConfig`).
2. Run bare create with `["--no-config"]` only (no UX flags, no task).
3. Expect create success under default WRK_HOME layout; all UX mocks stay silent.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeFullCreateUXConfig(t, req.WrkHome)
	req.Args = []string{"--no-config"}
	return nil
}
```
