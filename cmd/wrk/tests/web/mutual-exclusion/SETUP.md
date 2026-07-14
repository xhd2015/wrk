# Scenario

**Feature**: wrk --web is mutually exclusive with other modes

```
wrk --web + another mode flag (--list, --status, …)
  -> non-zero exit
  -> stderr mentions mutual exclusion
  -> stdout empty
```

## Steps

- Descendants combine `--web` with another mode flag.

## Context

- Prefer a single clear error in the same family as `--main` / `--projects`
  (e.g. `wrk: --web is mutually exclusive with other modes`).

```go
func Setup(t *testing.T, req *Request) error {
	// Error path: run wrk to completion (not the long-running WebProbe path).
	req.WebProbe = false
	req.WebPath = ""
	if req.RepoDir == "" {
		req.RepoDir = req.WorkRoot
	}
	return nil
}
```
