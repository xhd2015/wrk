# Scenario

**Feature**: CLI one-shot UX flags still run with `--no-config` (config not base-merged)

```
# full create.* config on disk
wrk --no-config -t 'ship feature' --new-window --new-terminal --open-in-agent
  -> create under WRK_HOME with task slug
  -> space once
  -> iterm ForceNew + agent follow-up
  -> outer agent-run NOT exec'd
  -> UX driven solely by CLI flags (config would be redundant here but must not
     be required and must not double-apply)
```

## Steps

1. Write full create UX config (`writeFullCreateUXConfig`) so config would fire without `--no-config`.
2. Run with `--no-config` + task + full CLI UX flag set.
3. Expect flag-driven pipeline identical to `pipeline/flags/full-pipeline`.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	writeFullCreateUXConfig(t, req.WrkHome)
	req.TaskDesc = "ship feature"
	req.TaskFlag = "-t"
	req.Args = []string{"--no-config", "--new-window", "--new-terminal", "--open-in-agent"}
	return nil
}
```
