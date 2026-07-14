# Scenario

**Feature**: wrk --projects perf log emits structured latency events

```
WRK_PROJECTS_PERF_LOG set -> JSONL events without changing stdout
```

```go
func Setup(t *testing.T, req *Request) error {
	ensurePerfProfileHelpersUsed()
	return nil
}
```