# Scenario

**Feature**: successful wrk --main --cd records events.jsonl command "cd"

```
WRK_FOLLOWUP_FILE set
linked-wt -> wrk --main --cd
  -> last event: command=cd, exit_code=0, args include --main and --cd
```

## Steps

1. Create main + linked worktree; open channel.
2. Run `wrk --main --cd`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo, linkedWT := initLinkedWorktree(t, req)
	req.MainRepo = mainRepo
	req.RepoDir = linkedWT
	enableFollowupChannel(t, req)
	setMainCdArgs(req, "--main", "--cd")
	return nil
}
```
