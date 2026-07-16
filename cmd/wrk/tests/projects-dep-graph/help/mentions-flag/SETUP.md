# Scenario

**Feature**: wrk -h mentions --projects-dep-graph

```
workspace/ -> wrk -h -> usage text includes --projects-dep-graph
```

## Steps

1. Run `wrk -h` from neutral cwd.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"-h"}
	return nil
}
```
