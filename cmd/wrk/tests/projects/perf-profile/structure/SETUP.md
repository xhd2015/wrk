# Scenario

**Feature**: wrk --projects perf log exposes structural inefficiencies

```
duplicate ListLinked calls -> single shared list per project
```

```go
func Setup(t *testing.T, req *Request) error {
	ensurePerfProfileHelpersUsed()
	return nil
}
```