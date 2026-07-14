# Scenario

**Feature**: dry-run detects outdated script and reports would update

```
pre-seed outdated bash.sh + both profile markers
wrk --bash-integration --install --dry-run
  -> bash integration: would update
  -> script (would update); markers (marker is up to date)
  -> no filesystem writes
```

## Steps

1. Pre-seed outdated `bash.sh` and both markers.
2. Run install dry-run.

```go
func Setup(t *testing.T, req *Request) error {
	requireMode(t, req, "install")
	if !req.DryRun {
		t.Fatalf("expected dry-run install")
	}
	req.PreExistingBashSh = outdatedBashShContent()
	req.PreExistingBashProfile = preInstalledProfileContent()
	req.PreExistingBashRC = preInstalledProfileContent()
	return nil
}
```
