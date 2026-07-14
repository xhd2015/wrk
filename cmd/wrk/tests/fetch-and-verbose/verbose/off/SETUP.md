# Scenario

**Feature**: wrk without -v produces empty stderr on healthy runs

```
no -v flag -> zero stderr logging overhead
```

```go
func Setup(t *testing.T, req *Request) error {
	skipIfNoGit(t)
	return nil
}
```