# Scenario

**Feature**: wrk --projects perf log exposes structural inefficiencies

```
duplicate ListLinked calls -> single shared list per project
```

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	ensurePerfProfileHelpersUsed()
	return nil
}
```