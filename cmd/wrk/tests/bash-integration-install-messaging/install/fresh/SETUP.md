# Scenario

**Feature**: fresh install reports installed for summary, script, and markers

```
empty fake HOME + WRK_HOME
wrk --bash-integration --install
  -> bash integration: installed
  -> script (installed); both markers (marker installed)
```

## Steps

1. Run install with no pre-seeded script or profile markers.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	requireMode(t, req, "install")
	if req.DryRun {
		t.Fatalf("expected real install, not dry-run")
	}
	requireNoPreseed(t, req)
	return nil
}
```
