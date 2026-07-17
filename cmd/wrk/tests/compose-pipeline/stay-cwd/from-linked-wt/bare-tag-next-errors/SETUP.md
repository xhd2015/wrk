# Scenario

**Feature**: bare `--tag-next` from linked worktree errors (activeRoot is WT, not main)

```
# Must NOT silently resolve/tag main from a linked worktree
linked wt -> wrk --tag-next
  -> non-zero
  -> stderr names --tag-next and main repository requirement
```

## Steps

1. Linked worktree ahead of main.
2. Run bare `--tag-next` with cwd = linked wt.

```go
func Setup(t *testing.T, req *Request) error {
	setupAPLinkedAhead(t, req)
	req.Args = []string{"--tag-next"}
	return nil
}
```
