# Scenario

**Feature**: `wrk --new --status` is mutually exclusive

```
wrk --new --status -> non-zero; mutually exclusive; no worktree created
```

## Steps

1. Init main repo (parent).
2. Run `wrk --new --status`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--new", "--status"}
	return nil
}
```
