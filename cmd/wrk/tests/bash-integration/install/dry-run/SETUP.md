# Scenario

**Feature**: install dry-run previews changes without writing

```
wrk --bash-integration --install --dry-run -> planned stdout only
```

## Steps

1. Set `req.Mode = "install"` and `req.DryRun = true`.

```go
func Setup(t *testing.T, req *Request) error {
	req.Mode = "install"
	req.DryRun = true
	return nil
}
```