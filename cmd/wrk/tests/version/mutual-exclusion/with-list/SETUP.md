# Scenario

**Feature**: wrk --version --list is mutually exclusive

```
workspace/ -> wrk --version --list -> non-zero, mutually exclusive
```

## Steps

1. Run `wrk --version --list` from neutral cwd.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	req.Args = []string{"--version", "--list"}
	return nil
}
```