# Scenario

**Feature**: in-place wrk --main --cd from linked worktree writes follow-up to main

```
WRK_FOLLOWUP_FILE set
linked-wt -> wrk --main --cd
  -> empty stdout; follow-up: cd <mainRepo>\n
```

## Steps

1. Create main + linked worktree; open follow-up channel (parent).
2. cwd = linked worktree root; Args = `--main`, `--cd`.

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
