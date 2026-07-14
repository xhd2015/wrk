# Scenario

**Feature**: uninstall dry-run previews marker removal

```
wrk --bash-integration --uninstall --dry-run -> preview stdout only
```

## Steps

1. Set `req.DryRun = true`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Mode = "uninstall"
	req.DryRun = true
	return nil
}
```