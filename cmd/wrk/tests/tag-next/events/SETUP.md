# Scenario

**Feature**: wrk --tag-next appends events.jsonl

```
# successful tag-next -> events.jsonl command=tag-next
wrk --tag-next -> event logged under WRK_HOME
```

## Preconditions

- Auto-record and event logging run on every wrk invocation.

## Steps

- Descendants run a successful `--tag-next` and assert the last event.

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	tagNextEnsureHelpersUsed()
	return nil
}
```