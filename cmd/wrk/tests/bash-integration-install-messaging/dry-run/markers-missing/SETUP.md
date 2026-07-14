# Scenario

**Feature**: dry-run with current script and missing markers reports would update

```
pre-seed current embedded bash.sh only (no profile markers)
wrk --bash-integration --install --dry-run
  -> bash integration: would update
  -> script (is up to date); markers (marker would install)
  -> no filesystem writes
```

## Steps

1. Seed current embedded `bash.sh` (`SeedCurrentScript`).
2. Leave profiles empty.
3. Run install dry-run.

```go
func Setup(t *testing.T, req *Request) error {
	requireMode(t, req, "install")
	if !req.DryRun {
		t.Fatalf("expected dry-run install")
	}
	req.SeedCurrentScript = true
	return nil
}
```
