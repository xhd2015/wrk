# Scenario

**Feature**: wrk --cd is mutually exclusive with other modes and with --no-cd

```
wrk --cd <path> --list   -> non-zero, mutually exclusive
wrk --cd <path> --no-cd  -> non-zero (exclusive with --no-cd)
wrk --cd <path> --where x -> non-zero, mutually exclusive
```

## Preconditions

- Target path may exist; rejection happens at mode selection before shell/follow-up.

## Steps

- Descendants set Args combining `--cd` with another mode/flag.

## Context

- Error stderr should mention mutually exclusive (or specifically `--cd` / `--no-cd`).

```go
func Setup(t *testing.T, req *Request) error {
	target := cdAbsTarget(t, req, "jumpto")
	req.MainRepo = target
	req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	return nil
}
```
