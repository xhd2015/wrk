# Scenario

**Feature**: --sync and --done are mutually exclusive

```
# wrk --sync --done -> non-zero, stderr mutually exclusive
wrk --sync --done -> error before sync body
```

## Steps

1. `initMainOnlyRepo` (valid git cwd so mode clash fires first).
2. Run `wrk --sync --done` from the main repo.

```go
func Setup(t *testing.T, req *Request) error {
	initMainOnlyRepo(t, req)
	req.Args = []string{"--sync", "--done"}
	return nil
}
```
