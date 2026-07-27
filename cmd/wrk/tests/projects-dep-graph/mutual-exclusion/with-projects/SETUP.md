# Scenario

**Feature**: wrk --projects-dep-graph --projects is mutually exclusive

```
workspace/ -> wrk --projects-dep-graph --projects
  -> non-zero, mutually exclusive
```

## Steps

1. Run both exclusive project-related flags together.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--projects-dep-graph", "--projects"}
	return nil
}
```
