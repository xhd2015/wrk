# Scenario

**Feature**: `wrk --new --list` is mutually exclusive

```
wrk --new --list -> non-zero; mutually exclusive; no worktree created
```

## Steps

1. Init main repo (parent).
2. Run `wrk --new --list`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--new", "--list"}
	return nil
}
```
