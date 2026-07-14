# Scenario

**Feature**: wrk --bash-integration --uninstall strips profile markers and keeps bash.sh

```
installed state
wrk --bash-integration --uninstall -> markers removed from both profiles; bash.sh remains
```

## Steps

1. Set `req.Mode = "uninstall"`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Mode = "uninstall"
	req.DryRun = false
	return nil
}
```