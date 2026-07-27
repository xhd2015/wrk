# Scenario

**Feature**: wrk --projects perf log emits structured latency events

```
WRK_PROJECTS_PERF_LOG set -> JSONL events without changing stdout
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensurePerfProfileHelpersUsed()
	return nil
}
```