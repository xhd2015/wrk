# Scenario

**Feature**: uninstall dry-run without markers reports already uninstalled

```
empty fake HOME
wrk --bash-integration --uninstall --dry-run -> already uninstalled
```

## Steps

1. Run uninstall dry-run on empty fake HOME.

```go
func Setup(t *testing.T, req *Request) error {
	requireMode(t, req, "uninstall")
	if !req.DryRun {
		t.Fatalf("expected dry-run uninstall")
	}
	requireNoPreseed(t, req)
	return nil
}
```