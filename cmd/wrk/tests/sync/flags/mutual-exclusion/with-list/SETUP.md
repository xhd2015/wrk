# Scenario

**Feature**: --sync and --list are mutually exclusive

```
# wrk --sync --list -> non-zero, stderr mutually exclusive
wrk --sync --list -> error before sync body
```

## Steps

1. `initMainOnlyRepo`.
2. Run `wrk --sync --list` from the main repo.

```go
func Setup(t *testing.T, req *Request) error {
	initMainOnlyRepo(t, req)
	req.Args = []string{"--sync", "--list"}
	return nil
}
```
