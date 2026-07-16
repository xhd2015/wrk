# Scenario

**Feature**: wrk --gen-commit-msg --sync (no primary) is mutually exclusive

```
workspace/ -> wrk --gen-commit-msg --sync
  -> non-zero; mutually exclusive; empty stdout
# P2 allows --sync only as post stage after --done/--merge-back with gen-commit pre
```

## Steps

1. Run `wrk --gen-commit-msg --sync` from neutral cwd (no git required).

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--gen-commit-msg", "--sync"}
	return nil
}
```
