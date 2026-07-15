# Scenario

**Feature**: bare wrk --dry-run rejected unless paired with --all-deps or --tag-next

```
# wrk --dry-run alone -> error mentioning --all-deps and --tag-next
wrk --dry-run -> validation error before tagscope
```

## Steps

1. `setupRootBumpRepo` (valid git cwd).
2. Run `wrk --dry-run` without `--tag-next` or `--all-deps`.

```go
func Setup(t *testing.T, req *Request) error {
	setupRootBumpRepo(t, req)
	req.Args = []string{"--dry-run"}
	return nil
}
```