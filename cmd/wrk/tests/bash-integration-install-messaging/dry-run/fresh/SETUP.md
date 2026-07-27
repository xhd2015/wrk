# Scenario

**Feature**: dry-run on empty HOME reports would install without writes

```
empty fake HOME + WRK_HOME
wrk --bash-integration --install --dry-run
  -> bash integration: would install
  -> script (would install); markers (marker would install)
  -> no files created
```

## Steps

1. Run install dry-run with no pre-seeded state.

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
