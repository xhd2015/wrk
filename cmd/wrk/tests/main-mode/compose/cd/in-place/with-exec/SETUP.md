# Scenario

**Feature**: wrk --main --cd --exec allowed; follow-up + child in main dir

```
WRK_FOLLOWUP_FILE set
linked-wt -> wrk --main --cd --exec pwd
  -> follow-up: cd <main>\n
  -> stdout: <main>\n (from pwd child)
  -> exit 0
```

## Steps

1. Create main + linked worktree; open channel.
2. Args = `--main`, `--cd`, `--exec`, `pwd`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	mainRepo, linkedWT := initLinkedWorktree(t, req)
	req.MainRepo = mainRepo
	req.RepoDir = linkedWT
	enableFollowupChannel(t, req)
	req.Args = []string{"--main", "--cd", "--exec", "pwd"}
	req.TargetDir = ""
	return nil
}
```
