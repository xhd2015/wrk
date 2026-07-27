# Scenario

**Feature**: parent-shell follow-up cd is skipped when create opens agent and/or terminal

```
# shell cwd = FakeHome + WRK_FOLLOWUP_FILE open
wrk <mainRepo> --open-in-agent           -> worktree ok; follow-up empty
wrk <mainRepo> --new-terminal            -> worktree ok; follow-up empty
wrk <mainRepo>                           -> follow-up: cd <worktree>  (regression)
wrk <mainRepo> --force-cd --open-in-agent -> follow-up: cd <worktree>
```

## Preconditions

- Uses create-ux mocks (fake agent-run / iterm) so UX flags do not call real tools.
- `HOME=FakeHome` and process cwd (`RepoDir`) = FakeHome so home gate would otherwise open.
- `WRK_FOLLOWUP_FILE` prepared via root `UseFollowupEnv` + `FollowupFile`.

## Steps

1. `setupCreateUXFromFakeHome` seeds main repo, mocks, FakeHome cwd, and follow-up channel.
2. Leaves add UX / force-cd flags and assert follow-up file content.

```go
import (
	"testing"
	"github.com/xhd2015/doctest/session"
)

func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	return nil
}
```
