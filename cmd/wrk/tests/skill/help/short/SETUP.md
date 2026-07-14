# Scenario

**Feature**: wrk skill -h prints skill-level usage

```
workspace/ -> wrk skill -h -> usage on stdout, exit 0
```

## Steps

1. Run `wrk skill -h` from neutral cwd.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"skill", "-h"}
	return nil
}
```
