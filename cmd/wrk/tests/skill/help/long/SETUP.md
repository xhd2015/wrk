# Scenario

**Feature**: wrk skill --help prints skill-level usage

```
workspace/ -> wrk skill --help -> usage on stdout, exit 0
```

## Steps

1. Run `wrk skill --help` from neutral cwd.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"skill", "--help"}
	return nil
}
```
