# Scenario

**Feature**: `wrk --new --done` is mutually exclusive

```
wrk --new --done -> non-zero; mutually exclusive; no worktree created
```

## Steps

1. Init main repo (parent).
2. Run `wrk --new --done`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--new", "--done"}
	return nil
}
```
