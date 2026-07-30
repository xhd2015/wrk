# Scenario

**Feature**: wrk --main error paths — non-git cwd, mutual exclusion, unexpected arguments

```
# invalid invocation or environment
wrk --main (not git)              -> non-zero; not a git repository
wrk --main --list                 -> non-zero; mutually exclusive
wrk --main --where --cd           -> non-zero; mutually exclusive (real exclusives)
wrk --main some-path              -> non-zero; unexpected arguments
# note: --main --cd / --main --where are valid compose (see compose/); extra path → unexpected args
```

## Preconditions

- Rejection for mutual exclusion / unexpected args happens at mode selection
  (before shell launch); fake bash is not required for error leaves.

## Steps

1. Descendants set cwd and Args combining `--main` with invalid extras or non-git cwd.
2. Run wrk; assert non-zero exit and stderr.

## Context

- No nested shell on error paths.
- Compose success and compose arity rejects live under `compose/`, not here.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	// Error leaves configure RepoDir / Args themselves.
	// Ensure helpers from parent main/SETUP.md stay referenced.
	ensureMainHelpersUsed()
	return nil
}
```
