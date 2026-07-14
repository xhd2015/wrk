# Scenario

**Feature**: full create.* config is silently ignored when `<target-dir>` is set

```
# config window+terminal+agent all on; no CLI UX flags
wrk <myrepo> <target-dir>
  -> exit 0; stdout = abs worktree path\n
  -> no space.CreateAndActivate
  -> no iterm OpenConfig
  -> no agent-run (outer or follow-up)
```

## Steps

1. Write full create UX config (`writeFullCreateUXConfig`).
2. Run `wrk <mainRepo> <spawnTarget>` with no UX flags (group Setup already set
   TargetDir/SpawnDir).
3. Expect create success at exact spawn path; all UX mocks stay silent.

```go
import (
	"testing"
)

func Setup(t *testing.T, req *Request) error {
	writeFullCreateUXConfig(t, req.WrkHome)
	req.Args = nil
	return nil
}
```
