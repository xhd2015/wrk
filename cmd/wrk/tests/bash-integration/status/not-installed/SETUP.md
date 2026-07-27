# Scenario

**Feature**: status reports not installed on empty env

```
empty fake HOME + empty WRK_HOME integration
wrk --bash-integration --status -> not installed, exit 1
```

## Steps

1. Run status with no pre-seeded integration files.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	requireMode(t, req, "status")
	requireNoPreseed(t, req)
	return nil
}
```