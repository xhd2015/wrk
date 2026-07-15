# Scenario

**Feature**: --sync and --status are mutually exclusive

```
# wrk --sync --status -> non-zero, stderr mutually exclusive
wrk --sync --status -> error before sync body
```

## Steps

1. `initMainOnlyRepo`.
2. Run `wrk --sync --status` from the main repo.

```go
func Setup(t *testing.T, req *Request) error {
	initMainOnlyRepo(t, req)
	req.Args = []string{"--sync", "--status"}
	return nil
}
```
