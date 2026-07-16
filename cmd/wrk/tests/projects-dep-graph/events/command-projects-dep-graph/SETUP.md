# Scenario

**Feature**: successful --projects-dep-graph records events.jsonl command "projects-dep-graph"

```
# empty registry success -> events.jsonl last event command=projects-dep-graph
(no projects.json) -> wrk --projects-dep-graph -> event appended
```

## Steps

1. Leave `projects.json` absent under isolated WRK_HOME.
2. Run `wrk --projects-dep-graph` from the neutral workspace cwd.

```go
func Setup(t *testing.T, req *Request) error {
	req.Args = []string{"--projects-dep-graph"}
	return nil
}
```
