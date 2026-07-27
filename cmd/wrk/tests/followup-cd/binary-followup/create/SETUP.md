# Scenario

**Feature**: create mode follow-up behavior (home-gated default; target-dir never)

```
# shell cwd == user home (FakeHome) + WRK_FOLLOWUP_FILE + default create
wrk <mainRepo> -> follow-up: cd <new-worktree>

# shell cwd == main repo (not home) + WRK_FOLLOWUP_FILE
wrk -> worktree created; follow-up empty

# explicit second positional <target-dir> (missing or existing parent)
# even from FakeHome with WRK_FOLLOWUP_FILE set → follow-up empty (no auto-cd)
# unless --force-cd is also set
wrk <mainRepo> <target> -> worktree at/under target; follow-up empty

# --force-cd bypasses home gate (channel open or interactive fallback)
cwd=main; wrk --force-cd + WRK_FOLLOWUP_FILE -> cd <worktree>
cwd=main; wrk --force-cd (no channel) -> install hint + shell @ worktree

# --no-cd or unset env always suppress write (without --force-cd)
```

## Steps

1. Descendants seed main repo and choose shell cwd (FakeHome vs mainRepo).
2. Target-dir leaves pass a second absolute positional path and assert empty follow-up.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	requireMode(t, req, "binary")
	return nil
}
```
