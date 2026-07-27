# Scenario

**Feature**: re-install when fully current reports is up to date

```
pre-install full integration (current script + both markers)
wrk --bash-integration --install
  -> bash integration: is up to date
  -> all components (is up to date)
```

## Steps

1. Pre-install via `req.PreInstall = true` so script and markers match current embedded content.
2. Run install again (the measured run).

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	requireMode(t, req, "install")
	if req.DryRun {
		t.Fatalf("expected real install, not dry-run")
	}
	req.PreInstall = true
	return nil
}
```
