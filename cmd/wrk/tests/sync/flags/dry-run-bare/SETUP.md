# Scenario

**Feature**: bare wrk --dry-run rejected unless paired with an allowed mode (incl. --sync)

```
# wrk --dry-run alone -> error listing all dry-run hosts incl. --sync and --propagate-tags
wrk --dry-run -> validation error before any mode body
```

## Steps

1. `initMainOnlyRepo` (valid git cwd).
2. Run `wrk --dry-run` without any dry-run host mode.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	initMainOnlyRepo(t, req)
	req.Args = []string{"--dry-run"}
	return nil
}
```
