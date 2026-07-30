# Scenario

**Feature**: bare wrk --dry-run rejected without a dry-run host mode

```
# wrk --dry-run alone -> error listing done, merge-back, all-deps, tag-next, propagate-tags, sync
wrk --dry-run -> validation error before tagscope
```

## Steps

1. `setupRootBumpRepo` (valid git cwd).
2. Run `wrk --dry-run` without any host (`--done` / `--merge-back` / `--tag-next` / `--propagate-tags` / `--sync`).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	setupRootBumpRepo(t, req)
	req.Args = []string{"--dry-run"}
	return nil
}
```
