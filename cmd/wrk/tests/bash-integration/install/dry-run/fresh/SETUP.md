# Scenario

**Feature**: install dry-run on empty HOME previews script and dual profile markers

```
empty fake HOME
wrk --bash-integration --install --dry-run -> preview stdout, no filesystem writes
```

## Steps

1. Run install dry-run on empty fake HOME.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	requireMode(t, req, "install")
	if !req.DryRun {
		t.Fatalf("expected dry-run install")
	}
	requireNoPreseed(t, req)
	return nil
}
```