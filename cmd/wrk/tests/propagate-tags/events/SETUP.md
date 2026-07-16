# Scenario

**Feature**: wrk --propagate-tags appends events.jsonl with command "propagate-tags"

```
# successful bare --propagate-tags -> events.jsonl last event command=propagate-tags
source repo -> wrk --propagate-tags --dry-run -> event logged under WRK_HOME
```

## Preconditions

- Auto-record and event logging run on every wrk invocation.
- Event command identity for bare `--propagate-tags` is `propagate-tags` (not tag-next).

## Steps

- Descendants run a successful bare `--propagate-tags` and assert the last event.

```go
func Setup(t *testing.T, req *Request) error {
	propTagsEnsureHelpersUsed()
	return nil
}
```
