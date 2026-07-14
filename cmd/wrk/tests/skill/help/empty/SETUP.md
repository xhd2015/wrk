# Scenario

**Feature**: wrk skill with no flags prints skill-level usage

```
workspace/ -> wrk skill -> usage on stdout, exit 0
```

## Steps

1. Run `wrk skill` with no further args from neutral cwd.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"skill"}
	return nil
}
```
