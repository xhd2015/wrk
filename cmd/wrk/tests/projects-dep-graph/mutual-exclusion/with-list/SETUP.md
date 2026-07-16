# Scenario

**Feature**: wrk --projects-dep-graph --list is mutually exclusive

```
workspace/ -> wrk --projects-dep-graph --list
  -> non-zero, mutually exclusive
```

## Steps

1. Run `--projects-dep-graph` with `--list`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--projects-dep-graph", "--list"}
	return nil
}
```
