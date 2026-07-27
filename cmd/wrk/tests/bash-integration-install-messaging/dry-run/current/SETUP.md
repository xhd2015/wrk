# Scenario

**Feature**: dry-run when fully current reports is up to date

```
pre-install full integration (current script + both markers)
wrk --bash-integration --install --dry-run
  -> bash integration: is up to date
  -> all components is up to date
  -> no filesystem writes
```

## Steps

1. Pre-install via `req.PreInstall = true`.
2. Run install dry-run.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	requireMode(t, req, "install")
	if !req.DryRun {
		t.Fatalf("expected dry-run install")
	}
	req.PreInstall = true
	return nil
}
```
