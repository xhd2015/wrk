# Scenario

**Feature**: wrk --gen-commit-msg --status is mutually exclusive

```
workspace/ -> wrk --gen-commit-msg --status
  -> non-zero; mutually exclusive; empty stdout
```

## Steps

1. Run `wrk --gen-commit-msg --status` from neutral cwd (no git required).

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--gen-commit-msg", "--status"}
	return nil
}
```
