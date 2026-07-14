# Scenario

**Feature**: wrk --rm is idempotent when entry absent

```
wrk --rm <path> (not in projects.json) -> exit 0, empty stdout, projects.json unchanged
wrk --rm twice -> second call exit 0, empty stdout
```

## Steps

- Descendants vary whether the path was never recorded or already removed.

```go
func Setup(t *testing.T, req *Request) error {
	ensureRemoveHelpersUsed()
	return nil
}
```
