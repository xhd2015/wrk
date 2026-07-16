# Scenario

**Feature**: wrk --tag-next appends events.jsonl (primary command tag-next)

```
# successful tag-next (bare or with --propagate-tags) -> events.jsonl command=tag-next
wrk --tag-next [--propagate-tags] -> event logged under WRK_HOME
```

## Preconditions

- Auto-record and event logging run on every wrk invocation.
- When composed with `--propagate-tags`, primary event command remains `tag-next`.

## Steps

- Descendants run a successful `--tag-next` (optionally with `--propagate-tags`) and assert the last event.

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	tagNextEnsureHelpersUsed()
	return nil
}
```