# Scenario

**Feature**: wrk --bash-integration --status reports integration filesystem state

```
wrk --bash-integration --status -> installed | not installed | partial (read-only)
```

## Steps

1. Set `req.Mode = "status"`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Mode = "status"
	req.PreInstall = false
	return nil
}
```