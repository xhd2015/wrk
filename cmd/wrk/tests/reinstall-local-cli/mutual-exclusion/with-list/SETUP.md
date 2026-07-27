# Scenario

**Feature**: wrk --reinstall-local --list is mutually exclusive

```
# C4
workspace/ -> wrk --reinstall-local --list -> non-zero, mutually exclusive
```

## Steps

1. Run `wrk --reinstall-local --list` from neutral module dir (no go.mod required;
   exclusion should fire before planning).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	req.Args = []string{"--reinstall-local", "--list"}
	return nil
}
```
