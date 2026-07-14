# Scenario

**Feature**: wrk --bash-integration --list is mutually exclusive

```
wrk --bash-integration --list -> non-zero exit, stderr error, empty stdout
```

## Steps

1. Run `wrk --bash-integration --list` from neutral cwd.

```go
func Setup(t *testing.T, req *Request) error {
	requireMode(t, req, "mutual")
	return nil
}
```