# Scenario

**Feature**: `--open-in-agent` without terminal runs agent-run in current process with `--dir`

```
wrk -t 'ship feature' --open-in-agent
  -> create; agent-run argv includes --dir <abs-worktree>; no space; no iterm
  -> process cwd of agent-run need not equal worktree (--dir is source of truth)
```

## Steps

1. Run with task + `--open-in-agent`.

```go
func Setup(t *testing.T, req *Request) error {
	req.TaskDesc = "ship feature"
	req.TaskFlag = "-t"
	req.Args = []string{"--open-in-agent"}
	return nil
}
```
