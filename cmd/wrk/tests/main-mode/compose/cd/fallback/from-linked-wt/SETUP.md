# Scenario

**Feature**: fallback wrk --main --cd from linked wt prints path, warns, shells at main

```
WRK_FOLLOWUP_FILE unset
fake bash on PATH
linked-wt -> wrk --main --cd
  -> stdout main\n
  -> stderr mentions wrk --bash-integration --install
  -> fake shell cwd = main; exit 0
```

## Steps

1. Create main + linked worktree.
2. Install fake bash (exit 0); channel closed.
3. Run `wrk --main --cd` from linked worktree root.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo, linkedWT := initLinkedWorktree(t, req)
	req.MainRepo = mainRepo
	req.RepoDir = linkedWT
	installFakeBash(t, req, 0)
	// Channel closed: UseFollowupEnv left false.
	setMainCdArgs(req, "--main", "--cd")
	return nil
}
```
