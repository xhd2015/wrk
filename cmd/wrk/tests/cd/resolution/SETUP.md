# Scenario

**Feature**: wrk --cd path resolution errors and basename edge cases

```
# arg / path failures (no shell launch)
wrk --cd                    -> requires path argument
wrk --cd missing            -> does not exist
wrk --cd basename (0 match) -> does not exist
wrk --cd file               -> not a directory
wrk --cd ambiguous basename (non-TTY) -> error listing candidates
wrk --cd local-basename     -> uses cwd dir when present (no projects required)
```

## Preconditions

- Channel typically closed; failures must not launch interactive shell.
- Basename fallback uses `resolveDirArg(..., allowBasenameFallback=true, ...)`.

## Steps

- Descendants configure fixtures and Args / TargetDir.

## Context

- Extra positionals → `wrk: unexpected arguments` (covered under mutual-exclusion-adjacent
  or leaf-specific Args).

```go
func Setup(t *testing.T, req *Request) error {
	if req.RepoDir == "" {
		req.RepoDir = initNeutralCwd(t, req.WorkRoot, "workspace")
	}
	return nil
}
```
