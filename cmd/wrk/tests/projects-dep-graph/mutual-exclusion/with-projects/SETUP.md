# Scenario

**Feature**: wrk --projects-dep-graph --projects is mutually exclusive

```
workspace/ -> wrk --projects-dep-graph --projects
  -> non-zero, mutually exclusive
```

## Steps

1. Run both exclusive project-related flags together.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--projects-dep-graph", "--projects"}
	return nil
}
```
