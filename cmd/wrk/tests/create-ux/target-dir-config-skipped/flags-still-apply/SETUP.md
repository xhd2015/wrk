# Scenario

**Feature**: CLI one-shot UX flags still run with `<target-dir>` when config is empty

```
# empty / no create UX config
wrk <myrepo> <target-dir> -t 'ship feature' --new-window --new-terminal --open-in-agent
  -> create at target
  -> space once
  -> iterm ForceNew + agent follow-up
  -> outer agent-run NOT exec'd
```

## Steps

1. Leave create config empty/off (no `writeFullCreateUXConfig`).
2. Run with task + full CLI UX flag set and SpawnDir set by group Setup.
3. Expect flag-driven pipeline identical to `pipeline/flags/full-pipeline`, but
   worktree path is the spawn target.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	// Empty create UX: do not write create.* config (config.json may be absent).
	req.TaskDesc = "ship feature"
	req.TaskFlag = "-t"
	req.Args = []string{"--new-window", "--new-terminal", "--open-in-agent"}
	return nil
}
```
