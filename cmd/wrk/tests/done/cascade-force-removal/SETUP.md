# Scenario

**Bug**: non-TTY cascade must not force-remove ahead/diverged linked worktrees

```
# consumer wt + ahead external dep wt -> wrk --done (non-TTY)
consumer wt -> wrk --dep -> external wt ahead of dep main
wrk --done (no TTY, no -y) -> error; external wt + commits preserved (no force-remove)
```

## Preconditions

- Git and Go must be available.

## Steps

- Descendants build a consumer linked worktree with an external dependency worktree that has unpushed commits ahead of the dep main branch.
- Run `wrk --done` from the consumer worktree on a non-TTY stdin (default doctest pipe).

```go
import (
	"os/exec"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	if _, err := exec.LookPath("go"); err != nil {
		t.Skip("go not available")
	}
	return nil
}
```
