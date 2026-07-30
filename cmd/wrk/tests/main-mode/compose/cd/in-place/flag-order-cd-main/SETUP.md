# Scenario

**Feature**: wrk --cd --main (partner first) in-place success

```
WRK_FOLLOWUP_FILE set
linked-wt -> wrk --cd --main
  -> empty stdout; follow-up: cd <main>\n
```

## Steps

1. Create main + linked worktree; open channel.
2. Args = `--cd`, `--main`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo, linkedWT := initLinkedWorktree(t, req)
	req.MainRepo = mainRepo
	req.RepoDir = linkedWT
	enableFollowupChannel(t, req)
	setMainCdArgs(req, "--cd", "--main")
	return nil
}
```
