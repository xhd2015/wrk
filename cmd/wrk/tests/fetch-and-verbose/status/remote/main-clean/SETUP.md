# Scenario

**Feature**: tracked up-to-date main repo shows Remote: identical on root block

```
tracked main up-to-date -> wrk --status -> root Remote: identical
```

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}
```