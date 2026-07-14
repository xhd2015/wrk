# Scenario

**Feature**: wrk creates worktree from optional directory argument

```
# target dir is a git checkout; wrk invoked from WorkRoot
wrk <repoDir> -> same stdout + side effects as wrk from <repoDir>
```

## Preconditions

- Git must be available.

## Steps

- Tests run `wrk <repoDir>` with no extra args (`req.Args` empty).
- `req.TargetDir` points at the git repo; `req.RepoDir` stays `{WorkRoot}`.

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}
```