# Scenario

**Feature**: non-TTY cascade auto-yes merges/removes ahead linked worktrees (no hard guard)

```
# consumer wt + ahead external dep wt -> wrk --done (non-TTY)
# default auto-yes: cascade merges dep into dep main, removes external wt, then consumer done
```

## Preconditions

- Git and Go must be available.

## Steps

- Descendants build a consumer linked worktree with an external dependency worktree that has unpushed commits ahead of the dep main branch, drop the consumer local replace, and run `wrk --done` (optionally with `-y`) on non-TTY stdin.

```go
import (
	"os/exec"
)

func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not available")
	}
	return nil
}
```
