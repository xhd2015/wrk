# Scenario

**Feature**: wrk --tag-next --push publishes branch tip and new tags to origin

```
# Apply tags locally, then runPushMain(branch + created tags)
# Human stdout: tag-next block, blank line, pushed <branch> → origin/<branch>
wrk --tag-next --push -> local tags + origin branch tip + origin tag refs
wrk --tag-next --push --dry-run -> plan tags + would: git push for branch and tags
```

## Preconditions

- Repo has `origin` remote pointing at a bare repository.
- Target contract (Classic TDD): **branch + tags**, not tags-only via tagscope.Apply(Push).

## Steps

- Descendants use `setupPushRepo` and set `req.Args` for `--tag-next --push` (± `--dry-run`).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	skipIfNoGit(t)
	tagNextEnsureHelpersUsed()
	return nil
}
```
