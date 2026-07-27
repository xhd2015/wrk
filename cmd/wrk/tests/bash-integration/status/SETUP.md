# Scenario

**Feature**: wrk --bash-integration --status reports integration filesystem state

```
wrk --bash-integration --status -> installed | not installed | partial (read-only)
```

## Steps

1. Set `req.Mode = "status"`.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.Mode = "status"
	req.PreInstall = false
	return nil
}
```