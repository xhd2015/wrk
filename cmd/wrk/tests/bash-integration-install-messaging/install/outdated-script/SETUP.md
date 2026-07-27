# Scenario

**Feature**: install updates outdated bash.sh and reports updated

```
pre-seed outdated bash.sh + both profile markers
wrk --bash-integration --install
  -> bash integration: updated
  -> script (updated); markers (marker is up to date)
```

## Steps

1. Pre-seed outdated `bash.sh` and both profile markers.
2. Run install.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	requireMode(t, req, "install")
	if req.DryRun {
		t.Fatalf("expected real install, not dry-run")
	}
	req.PreExistingBashSh = outdatedBashShContent()
	req.PreExistingBashProfile = preInstalledProfileContent()
	req.PreExistingBashRC = preInstalledProfileContent()
	return nil
}
```
