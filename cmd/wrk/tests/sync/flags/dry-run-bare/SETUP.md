# Scenario

**Feature**: bare wrk --dry-run rejected unless paired with an allowed mode (incl. --sync)

```
# wrk --dry-run alone -> error mentioning --all-deps and --sync
wrk --dry-run -> validation error before any mode body
```

## Steps

1. `initMainOnlyRepo` (valid git cwd).
2. Run `wrk --dry-run` without `--sync`, `--all-deps`, or `--tag-next`.

```go
func Setup(t *testing.T, req *Request) error {
	initMainOnlyRepo(t, req)
	req.Args = []string{"--dry-run"}
	return nil
}
```
