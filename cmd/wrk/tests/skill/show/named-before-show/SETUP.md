# Scenario

**Feature**: wrk skill <name> --show accepts the name before the action flag

```
workspace/ -> wrk skill wrk-pipeline --show -> same stdout as --show wrk-pipeline
```

## Steps

1. Run `wrk skill wrk-pipeline --show` from neutral cwd.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"skill", "wrk-pipeline", "--show"}
	return nil
}
```
