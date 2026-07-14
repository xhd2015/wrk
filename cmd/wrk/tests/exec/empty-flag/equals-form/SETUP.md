# Scenario

**Feature**: equals form `--exec=value` is rejected

```
workspace/ -> wrk --exec=pwd -> non-zero; cut is not a value flag
```

## Steps

1. Run `wrk --exec=pwd` from WorkRoot (no git create needed if parse fails first).

```go
func Setup(t *testing.T, req *Request) error {
	req.RepoDir = req.WorkRoot
	req.Args = []string{"--exec=pwd"}
	return nil
}
```
