# Scenario

**Feature**: one missing marker with current script reports updated

```
pre-seed current bash.sh + marker only in .bashrc
wrk --bash-integration --install
  -> bash integration: updated
  -> script (is up to date)
  -> bash_profile (marker installed); bashrc (marker is up to date)
```

## Steps

1. Seed current embedded `bash.sh`.
2. Pre-seed marker only in `.bashrc` (leave `.bash_profile` without marker).
3. Run install.

```go
func Setup(t *testing.T, d *session.Doctest, req *Request) error {
	_ = d
	req.InProcess = true
	requireMode(t, req, "install")
	if req.DryRun {
		t.Fatalf("expected real install, not dry-run")
	}
	req.SeedCurrentScript = true
	req.PreExistingBashRC = preInstalledProfileContent()
	// .bash_profile intentionally unset / absent
	return nil
}
```
