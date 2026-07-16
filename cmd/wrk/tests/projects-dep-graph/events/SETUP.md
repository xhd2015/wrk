# Scenario

**Feature**: wrk --projects-dep-graph appends events.jsonl with command "projects-dep-graph"

```
# successful --projects-dep-graph -> events.jsonl last event command=projects-dep-graph
WRK_HOME -> wrk --projects-dep-graph -> event logged
```

## Preconditions

- Auto-record and event logging run on every wrk invocation.
- Event command identity for bare `--projects-dep-graph` is `projects-dep-graph`.

## Steps

- Descendants run a successful `--projects-dep-graph` and assert the last event.

```go
func Setup(t *testing.T, req *Request) error {
	depGraphEnsureHelpersUsed()
	return nil
}
```
